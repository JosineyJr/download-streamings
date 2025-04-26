package errors

import "errors"

var (
	ErrNotDefinedBearerToken = errors.New("bearer token not defined")
)
