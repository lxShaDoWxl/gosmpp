package coding

import (
	"github.com/linxGnu/gosmpp/coding/gsm7bit"
	"unicode"
)

// BestSafeCoding returns suitable encoding for a string.
// If string is ascii, then GSM7Bit. If not, then UCS2.
func BestSafeCoding(input string) Encoding {
	if len(gsm7bit.ValidateString(input)) == 0 {
		return GSM7BIT
	}
	return UCS2
}

// FindEncoding returns suitable encoding for a string.
// If string is ascii, then GSM7Bit. If not, then UCS2.
func FindEncoding(s string) (enc Encoding) {
	if isASCII(s) {
		enc = GSM7BIT
	} else {
		enc = UCS2
	}
	return
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
