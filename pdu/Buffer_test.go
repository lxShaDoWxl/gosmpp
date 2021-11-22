package pdu

import (
	"github.com/linxGnu/gosmpp/coding"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuffer(t *testing.T) {
	b := NewBuffer(nil)
	require.Nil(t, b.WriteCStringWithEnc("agjwklgjkwPץ", coding.HEBREW))
	require.Equal(t, "61676A776B6C676A6B7750F500", strings.ToUpper(b.HexDump()))
}
