package coding


type binary8bit2 struct{}

func (*binary8bit2) Encode(_ string) ([]byte, error) {
	return []byte{}, ErrNotImplEncode
}

func (*binary8bit2) Decode(_ []byte) (string, error) {
	return "", ErrNotImplDecode
}

func (*binary8bit2) DataCoding() byte { return BINARY8BIT2Coding }

