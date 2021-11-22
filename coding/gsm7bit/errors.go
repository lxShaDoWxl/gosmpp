package gsm7bit

import "errors"

// ErrInvalidCharacter means a given character can not be represented in GSM 7-bit encoding.
//
// This can only happen during encoding.
var ErrInvalidCharacter = errors.New("invalid gsm7 character")

// ErrInvalidByte means that a given byte is outside of the GSM 7-bit encoding range.
//
// This can only happen during decoding.
var ErrInvalidByte = errors.New("invalid gsm7 byte")


