package gsm7bit

import (
	"golang.org/x/text/transform"
	"math"
)

type gsm7Encoder struct {
	packed bool
}

func (g *gsm7Encoder) Reset() {
	/* no needed */
}

func (g *gsm7Encoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if len(src) == 0 {
		return 0, 0, nil
	}

	text := string(src) // work with []rune (a.k.a string) instead of []byte
	septets := make([]byte, 0, len(text))
	for _, r := range text {
		if v, ok := forwardLookup[r]; ok {
			septets = append(septets, v)
		} else if v, ok := forwardEscape[r]; ok {
			septets = append(septets, escapeSequence, v)
		} else {
			return 0, 0, ErrInvalidCharacter
		}
		nSrc++
	}

	nDst = len(septets)
	if g.packed {
		nDst = int(math.Ceil(float64(len(septets)) * 7 / 8))
	}
	if len(dst) < nDst {
		return 0, 0, transform.ErrShortDst
	}

	if !g.packed {
		copy(dst, septets)
		return nDst, nSrc, nil
	}

	nDst = pack(dst, septets)
	return
}
func pack(dst []byte, septets []byte) (nDst int) {
	nSeptet := 0
	for remain := len(septets); remain > 0; {
		// Pack by converting septets into octets.
		switch {
		case remain >= 8:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = (septets[nSeptet+2] & 0x7C >> 2) | (septets[nSeptet+3] & 0x07 << 5)
			dst[nDst+3] = (septets[nSeptet+3] & 0x78 >> 3) | (septets[nSeptet+4] & 0x0F << 4)
			dst[nDst+4] = (septets[nSeptet+4] & 0x70 >> 4) | (septets[nSeptet+5] & 0x1F << 3)
			dst[nDst+5] = (septets[nSeptet+5] & 0x60 >> 5) | (septets[nSeptet+6] & 0x3F << 2)
			dst[nDst+6] = (septets[nSeptet+6] & 0x40 >> 6) | (septets[nSeptet+7] & 0x7F << 1)
			nSeptet += 8
			nDst += 7
		case remain >= 7:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = (septets[nSeptet+2] & 0x7C >> 2) | (septets[nSeptet+3] & 0x07 << 5)
			dst[nDst+3] = (septets[nSeptet+3] & 0x78 >> 3) | (septets[nSeptet+4] & 0x0F << 4)
			dst[nDst+4] = (septets[nSeptet+4] & 0x70 >> 4) | (septets[nSeptet+5] & 0x1F << 3)
			dst[nDst+5] = (septets[nSeptet+5] & 0x60 >> 5) | (septets[nSeptet+6] & 0x3F << 2)
			dst[nDst+6] = septets[nSeptet+6] & 0x40 >> 6
			nSeptet += 7
			nDst += 7
		case remain >= 6:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = (septets[nSeptet+2] & 0x7C >> 2) | (septets[nSeptet+3] & 0x07 << 5)
			dst[nDst+3] = (septets[nSeptet+3] & 0x78 >> 3) | (septets[nSeptet+4] & 0x0F << 4)
			dst[nDst+4] = (septets[nSeptet+4] & 0x70 >> 4) | (septets[nSeptet+5] & 0x1F << 3)
			dst[nDst+5] = septets[nSeptet+5] & 0x60 >> 5
			nSeptet += 6
			nDst += 6
		case remain >= 5:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = (septets[nSeptet+2] & 0x7C >> 2) | (septets[nSeptet+3] & 0x07 << 5)
			dst[nDst+3] = (septets[nSeptet+3] & 0x78 >> 3) | (septets[nSeptet+4] & 0x0F << 4)
			dst[nDst+4] = septets[nSeptet+4] & 0x70 >> 4
			nSeptet += 5
			nDst += 5
		case remain >= 4:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = (septets[nSeptet+2] & 0x7C >> 2) | (septets[nSeptet+3] & 0x07 << 5)
			dst[nDst+3] = septets[nSeptet+3] & 0x78 >> 3
			nSeptet += 4
			nDst += 4
		case remain >= 3:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = (septets[nSeptet+1] & 0x7E >> 1) | (septets[nSeptet+2] & 0x03 << 6)
			dst[nDst+2] = septets[nSeptet+2] & 0x7C >> 2
			nSeptet += 3
			nDst += 3
		case remain >= 2:
			dst[nDst+0] = (septets[nSeptet+0] & 0x7F >> 0) | (septets[nSeptet+1] & 0x01 << 7)
			dst[nDst+1] = septets[nSeptet+1] & 0x7E >> 1
			nSeptet += 2
			nDst += 2
		case remain >= 1:
			dst[nDst+0] = septets[nSeptet+0] & 0x7F >> 0
			nSeptet++
			nDst++
		default:
			return
		}
		remain = len(septets) - nSeptet
	}
	return
}