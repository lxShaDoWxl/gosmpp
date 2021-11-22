package coding

import (
	"golang.org/x/text/encoding"
)

func encode(str string, encoder *encoding.Encoder) ([]byte, error) {
	return encoder.Bytes([]byte(str))
}

func decode(data []byte, decoder *encoding.Decoder) (st string, err error) {
	tmp, err := decoder.Bytes(data)
	if err == nil {
		st = string(tmp)
	}
	return
}

// FromDataCoding returns encoding from DataCoding value.
func FromDataCoding(code byte) (enc Encoding) {
	enc = codingMap[code]
	return
}
