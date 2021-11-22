package coding

import "golang.org/x/text/encoding/unicode"


type ucs2 struct{}

func (*ucs2) Encode(str string) ([]byte, error) {
	tmp := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	return encode(str, tmp.NewEncoder())
}

func (*ucs2) Decode(data []byte) (string, error) {
	tmp := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	return decode(data, tmp.NewDecoder())
}

func (*ucs2) ShouldSplit(text string, octetLimit uint) (shouldSplit bool) {
	return uint(len(text)*2) > octetLimit
}

func (c *ucs2) EncodeSplit(text string, octetLimit uint) (allSeg [][]byte, err error) {
	if octetLimit < 64 {
		octetLimit = 134
	}

	allSeg = [][]byte{}
	runeSlice := []rune(text)
	hextetLim := int(octetLimit / 2) // round down

	// hextet = 16 bits, the correct terms should be hexadectet
	fr, to := 0, hextetLim
	for fr < len(runeSlice) {
		if to > len(runeSlice) {
			to = len(runeSlice)
		}

		seg, err := c.Encode(string(runeSlice[fr:to]))
		if err != nil {
			return nil, err
		}
		allSeg = append(allSeg, seg)

		fr, to = to, to+hextetLim
	}

	return
}

func (*ucs2) DataCoding() byte { return UCS2Coding }
