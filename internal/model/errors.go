package model

import "errors"

// Errors
var (
	ErrUnsupportedEncoding = errors.New("unsupported encoding")
	ErrInvalidBase58       = errors.New("invalid base58 string")
	ErrInvalidBase85       = errors.New("invalid base85 string")
	ErrEmptyInput          = errors.New("empty input")
)
