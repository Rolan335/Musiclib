package controller

import "errors"

var ErrBadRequest = errors.New("bad request")
var ErrGatewayTimeout = errors.New("gateway timeout")
var ErrBadGateway = errors.New("bad gateway")
var ErrNotFound = errors.New("not found")
var ErrInternalServer = errors.New("internal server error")
var ErrFailedToParse = errors.New("failed to parse body")
