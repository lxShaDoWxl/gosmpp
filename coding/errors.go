package coding

import "fmt"

var (
	// ErrNotImplSplitterInterface indicates that encoding does not support Splitter interface
	ErrNotImplSplitterInterface = fmt.Errorf("Encoding not implementing Splitter interface")
	// ErrNotImplDecode indicates that encoding does not support Decode method
	ErrNotImplDecode = fmt.Errorf("Decode is not implemented in this Encoding")
	// ErrNotImplEncode indicates that encoding does not support Encode method
	ErrNotImplEncode = fmt.Errorf("Encode is not implemented in this Encoding")
)
