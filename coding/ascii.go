package coding


type ascii struct{}

func (*ascii) Encode(str string) ([]byte, error) {
	return []byte(str), nil
}

func (*ascii) Decode(data []byte) (string, error) {
	return string(data), nil
}

func (*ascii) DataCoding() byte { return ASCIICoding }

