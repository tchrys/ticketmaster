package main

const InvalidTokenHttpCode = 400

type InvalidTokenError struct {
	code int
	msg  string
}

func NewInvalidTokenError() *InvalidTokenError {
	return &InvalidTokenError{InvalidTokenHttpCode, "Invalid token"}
}

func (e *InvalidTokenError) Error() string {
	return e.msg
}
