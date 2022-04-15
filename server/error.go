package server

var (
	ErrPanic                       = NewError("panic")
	ErrResponseWrite               = NewError("response write")
	ErrRequestRead                 = NewError("request read")
	ErrRemoteConnect               = NewError("remote connect")
	ErrNotSupportHijacking         = NewError("hijacking not supported")
	ErrTLSSignHost                 = NewError("TLS sign host")
	ErrTLSHandshake                = NewError("TLS handshake")
	ErrAbsURLAfterCONNECT          = NewError("absolute URL after CONNECT")
	ErrRoundTrip                   = NewError("round trip")
	ErrUnsupportedTransferEncoding = NewError("unsupported transfer encoding")
	ErrNotSupportHTTPVer           = NewError("http version not supported")
)

// Error struct is base of library specific errors.
type Error struct {
	ErrString string
}

// NewError returns a new Error.
func NewError(errString string) *Error {
	return &Error{errString}
}

// Error implements error interface.
func (e *Error) Error() string {
	return e.ErrString
}
