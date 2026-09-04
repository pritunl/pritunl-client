package errortypes

import (
	"github.com/dropbox/godropbox/errors"
)

type UnknownError struct {
	errors.DropboxError
}

type NotFoundError struct {
	errors.DropboxError
}

type ReadError struct {
	errors.DropboxError
}

type WriteError struct {
	errors.DropboxError
}

type ParseError struct {
	errors.DropboxError
}

type ExecError struct {
	errors.DropboxError
}

type PermissionError struct {
	errors.DropboxError
}

type RequestError struct {
	errors.DropboxError
}
