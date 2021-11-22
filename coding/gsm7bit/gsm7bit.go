package gsm7bit

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// GSM7BITCoding is gsm-7bit coding
const GSM7BITCoding byte = 0x00

type EncDec struct {
	packed           bool
	encoder, decoder transform.Transformer
}

// NewEncDec returns a GSM 7-bit Bit Encoding.
//
// Set the packed flag to true if you wish to convert septets to octets,
// this should be false for most SMPP providers.
func NewEncDec(packed bool) *EncDec {
	return &EncDec{
		packed:  packed,
		encoder: &gsm7Encoder{packed: packed},
		decoder: &gsm7Decoder{packed: packed},
	}
}
func (c *EncDec) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: c.decoder}
}

func (c *EncDec) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: c.encoder}
}
func (c *EncDec) String() string {
	if c.packed {
		return "GSM 7-bit (Packed)"
	}
	return "GSM 7-bit (Unpacked)"
}
func (c *EncDec) Encode(str string) ([]byte, error) {
	return c.NewEncoder().Bytes([]byte(str))
}

func (c *EncDec) Decode(data []byte) (string, error) {
	tmp, err := c.NewDecoder().Bytes(data)
	if err != nil {
		return "", err
	}
	return string(tmp), nil
}

func (c *EncDec) DataCoding() byte { return GSM7BITCoding }

func (c *EncDec) ShouldSplit(text string, octetLimit uint) (shouldSplit bool) {
	return uint(len(text)) > octetLimit
}

func (c *EncDec) EncodeSplit(text string, octetLimit uint) (allSeg [][]byte, err error) {
	if octetLimit < 64 {
		octetLimit = 134
	}

	allSeg = [][]byte{}
	runeSlice := []rune(text)

	fr, to := 0, int(octetLimit)
	for fr < len(runeSlice) {
		if to > len(runeSlice) {
			to = len(runeSlice)
		}
		seg, err := c.Encode(string(runeSlice[fr:to]))
		if err != nil {
			return nil, err
		}
		allSeg = append(allSeg, seg)
		fr, to = to, to+int(octetLimit)
	}

	return
}

// ValidateString returns the characters, in the given text, that can not be represented in GSM 7-bit encoding.
func ValidateString(text string) []rune {
	invalidChars := make([]rune, 0, 4)
	for _, r := range text {
		if _, ok := forwardLookup[r]; !ok {
			if _, ok := forwardEscape[r]; !ok {
				invalidChars = append(invalidChars, r)
			}
		}
	}
	return invalidChars
}

// ValidateBuffer returns the bytes, in the given buffer, that are outside of the GSM 7-bit encoding range.
func ValidateBuffer(buffer []byte) []byte {
	invalidBytes := make([]byte, 0, 4)
	count := 0
	for count < len(buffer) {
		b := buffer[count]
		if b == escapeSequence {
			count++
			if count >= len(buffer) {
				invalidBytes = append(invalidBytes, b)
				break
			}
			e := buffer[count]
			if _, ok := reverseEscape[e]; !ok {
				invalidBytes = append(invalidBytes, b, e)
			}
		} else if _, ok := reverseLookup[b]; !ok {
			invalidBytes = append(invalidBytes, b)
		}
		count++
	}
	return invalidBytes
}
