package gosmpp

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/linxGnu/gosmpp/pdu"
	"golang.org/x/time/rate"
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
	aliveState  int32
}

func newTransceivable(conn *Connection, settings Settings) *transceivable {
	t := &transceivable{
		settings:    settings,
		conn:        conn,
		rateLimiter: settings.RateLimiter,
		ctx:         context.Background(),
		pending:     make(map[int32]func(pdu.PDU)),
	}

	t.out = newTransmittable(conn, Settings{
		WriteTimeout: settings.WriteTimeout,

		EnquireLink: settings.EnquireLink,

		OnSubmitError: settings.OnSubmitError,

		OnClosed: func(state State) {
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

	t.out.start()
	t.in.start()

	return t
}

// SystemID returns tagged SystemID which is attached with bind_resp from SMSC.
func (t *transceivable) SystemID() string {
	return t.conn.systemID
}

// Close transceiver and stop underlying daemons.
func (t *transceivable) Close() (err error) {
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
	err = t.rateLimit(ctx)
	if err != nil {
		return
	}
	sequence := p.GetSequenceNumber()
	returns := make(chan pdu.PDU, 1)
	t.pending[sequence] = func(resp pdu.PDU) { returns <- resp }

	defer delete(t.pending, sequence)

	err = t.out.Submit(p)
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
