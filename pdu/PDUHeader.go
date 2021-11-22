package pdu

import (
	"encoding/binary"
	"github.com/linxGnu/gosmpp/data"
	"math/rand"
)

func nextSequenceNumber() (v int32) {
	return rand.Int31()
}

// Header represents PDU header.
type Header struct {
	CommandLength  int32
	CommandID      data.CommandIDType
	CommandStatus  data.CommandStatusType
	SequenceNumber int32
}

// ParseHeader parses PDU header.
func ParseHeader(v [16]byte) (h Header) {
	h.CommandLength = int32(binary.BigEndian.Uint32(v[:]))
	h.CommandID = data.CommandIDType(binary.BigEndian.Uint32(v[4:]))
	h.CommandStatus = data.CommandStatusType(binary.BigEndian.Uint32(v[8:]))
	h.SequenceNumber = int32(binary.BigEndian.Uint32(v[12:]))
	return
}

// Unmarshal from buffer.
func (c *Header) Unmarshal(b *ByteBuffer) (err error) {
	var id, status int32
	c.CommandLength, err = b.ReadInt()
	if err == nil {
		id, err = b.ReadInt()
		if err == nil {
			c.CommandID = data.CommandIDType(id)
			if status, err = b.ReadInt(); err == nil {
				c.CommandStatus = data.CommandStatusType(status)
				c.SequenceNumber, err = b.ReadInt()
			}
		}
	}
	return
}


// AssignSequenceNumber assigns sequence random number.
func (c *Header) AssignSequenceNumber() {
	c.SetSequenceNumber(nextSequenceNumber())
}

// ResetSequenceNumber resets sequence number.
func (c *Header) ResetSequenceNumber() {
	c.SequenceNumber = 1
}

// GetSequenceNumber returns assigned sequence number.
func (c *Header) GetSequenceNumber() int32 {
	return c.SequenceNumber
}

// SetSequenceNumber manually sets sequence number.
func (c *Header) SetSequenceNumber(v int32) {
	c.SequenceNumber = v
}

// Marshal to buffer.
func (c *Header) Marshal(b *ByteBuffer) {
	b.Grow(16)
	b.WriteInt(c.CommandLength)
	b.WriteInt(int32(c.CommandID))
	b.WriteInt(int32(c.CommandStatus))
	b.WriteInt(c.SequenceNumber)
}
