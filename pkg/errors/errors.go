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
	ErrEmptyTarballPrefix              = errors.New("empty tarball prefix")
	ErrUnexpectedContentType           = errors.New("unexpected content type")
	ErrEmptySignature                  = errors.New("empty signature")
	ErrMissingSignatureDelimiter       = errors.New("missing signature delimiter")
	ErrUnexpectedSignatureLabel        = errors.New("unexpected signature label")
	ErrEmptyWebhookSecret              = errors.New("empty webhook secret")
)
