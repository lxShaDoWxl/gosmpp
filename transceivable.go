package gosmpp

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/linxGnu/gosmpp/pdu"
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
	"time"
)

type transceivable struct {
	settings Settings

	conn        *Connection
	in          *receivable
	out         *transmittable
	pending     map[int32]func(pdu.PDU)
	rateLimiter *rate.Limiter
	ctx         context.Context
	ctxCancel   context.CancelFunc
	aliveState  int32
	mutex       *sync.Mutex
}

func newTransceivable(conn *Connection, settings Settings) *transceivable {
	ctx, cancel := context.WithCancel(context.Background())
	t := &transceivable{
		settings:    settings,
		conn:        conn,
		rateLimiter: settings.RateLimiter,
		ctx:         ctx,
		ctxCancel:   cancel,
		pending:     make(map[int32]func(pdu.PDU)),
		mutex:       &sync.Mutex{},
	}

	t.out = newTransmittable(conn, Settings{
		WriteTimeout: settings.WriteTimeout,

		OnSubmitError: settings.OnSubmitError,

		OnClosed: func(state State) {
			defer cancel()
			switch state {
			case ExplicitClosing:
				return

			case ConnectionIssue:
				// also close input
				_ = t.in.close(ExplicitClosing)

				if t.settings.OnClosed != nil {
					t.settings.OnClosed(ConnectionIssue)
				}
			}
		},
	})

	t.in = newReceivable(conn, Settings{
		ReadTimeout: settings.ReadTimeout,

		OnPDU: t.onPDU(settings.OnPDU),

		OnReceivingError: settings.OnReceivingError,

		OnClosed: func(state State) {
			defer cancel()
			switch state {
			case ExplicitClosing:
				return

			case InvalidStreaming, UnbindClosing:
				// also close output
				_ = t.out.close(ExplicitClosing)

				if t.settings.OnClosed != nil {
					t.settings.OnClosed(state)
				}
			}

		},

		response: func(p pdu.PDU) {
			_ = t.Submit(p)
		},
	})

	t.start()
	return t
}

func (t *transceivable) start() {
	t.out.start()
	t.in.start()
	if t.settings.EnquireLink > 0 {
		go t.loopWithEnquireLink()
	}
}
func (t *transceivable) loopWithEnquireLink() {
	ticker := time.NewTicker(t.settings.EnquireLink)
	defer ticker.Stop()
	sendEnquireLink := func() {
		eqp := pdu.NewEnquireLink()
		ctxSubmit, cancel := context.WithTimeout(t.ctx, time.Minute*5)
		defer cancel()

		_, err := t.SubmitResp(ctxSubmit, eqp)
		if err != nil {
			t.settings.OnSubmitError(eqp, err)
			_ = t.Close()
			return
		}
	}
	for {
		select {
		case <-ticker.C:
			sendEnquireLink()

		case <-t.ctx.Done():
			return
		}
	}
}

// SystemID returns tagged SystemID which is attached with bind_resp from SMSC.
func (t *transceivable) SystemID() string {
	return t.conn.systemID
}

// Close transceiver and stop underlying daemons.
func (t *transceivable) Close() (err error) {
	defer t.ctxCancel()
	if atomic.CompareAndSwapInt32(&t.aliveState, Alive, Closed) {
		// closing input and output
		_ = t.out.close(StoppingProcessOnly)
		_ = t.in.close(StoppingProcessOnly)

		// close underlying conn
		err = t.conn.Close()

		// notify transceiver closed
		if t.settings.OnClosed != nil {
			t.settings.OnClosed(ExplicitClosing)
		}
	}
	return
}
func (t *transceivable) onPDU(cl PDUCallback) PDUCallback {
	return func(p pdu.PDU, responded bool) {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		if callback, ok := t.pending[p.GetSequenceNumber()]; ok {
			go callback(p)
		} else {
			if cl == nil {
				if p.CanResponse() {
					go func() {
						_ = t.Submit(p.GetResponse())
					}()
				}
			} else {
				go cl(p, responded)
			}

		}
	}
}

// Submit a PDU.
func (t *transceivable) Submit(p pdu.PDU) error {
	err := t.rateLimit(t.ctx)
	if err != nil {
		return err
	}
	return t.out.Submit(p)
}

// SubmitResp a PDU and response PDU.
func (t *transceivable) SubmitResp(ctx context.Context, p pdu.PDU) (resp pdu.PDU, err error) {
	if !p.CanResponse() {
		return nil, errors.New("Not response PDU")
	}
	sequence := p.GetSequenceNumber()
	returns := make(chan pdu.PDU, 1)

	func() {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		t.pending[sequence] = func(resp pdu.PDU) { returns <- resp }
	}()

	defer func() {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		delete(t.pending, sequence)
	}()

	err = t.Submit(p)
	if err != nil {
		return
	}
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case resp = <-returns:
	}
	return
}

func (t *transceivable) rateLimit(ctx context.Context) error {
	if t.rateLimiter != nil {
		ctxLimiter, cancelLimiter := context.WithTimeout(ctx, time.Minute)
		defer cancelLimiter()
		if err := t.rateLimiter.Wait(ctxLimiter); err != nil {
			return errors.Errorf("SMPP limiter failed: %v", err)
		}
	}
	return nil
}
