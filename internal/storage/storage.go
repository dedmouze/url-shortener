package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
	ErrAliasExists = errors.New("alias exists")
	ErrAppExists   = errors.New("app already exists")
	ErrAppNotFound = errors.New("app not found")
)
