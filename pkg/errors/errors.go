package errors

import "errors"

var (
	ErrEmptyOwner      = errors.New("empty owner")
	ErrEmptyName       = errors.New("empty repository")
	ErrEmptyBranch     = errors.New("empty branch")
	ErrEmptyToken      = errors.New("empty token")
	ErrNilReposBaseUrl = errors.New("nil repos base url")
)
