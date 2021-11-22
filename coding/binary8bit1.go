package coding

type binary8bit1 struct{}

func (*binary8bit1) Encode(_ string) ([]byte, error) {
	return []byte{}, ErrNotImplEncode
}

func (*binary8bit1) Decode(_ []byte) (string, error) {
	return "", ErrNotImplDecode
}

func (*binary8bit1) DataCoding() byte { return BINARY8BIT1Coding }
