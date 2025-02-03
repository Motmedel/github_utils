package errors

import "errors"

var (
	ErrNilReposBaseUrl                 = errors.New("nil repos base url")
	ErrEmptyOwner                      = errors.New("empty owner")
	ErrEmptyName                       = errors.New("empty repository")
	ErrEmptyBranch                     = errors.New("empty branch")
	ErrEmptyToken                      = errors.New("empty token")
	ErrNilTarballReader                = errors.New("nil tarball reader")
	ErrNilContentDisposition           = errors.New("nil content disposition")
	ErrEmptyContentDispositionFilename = errors.New("empty content disposition filename")
)
