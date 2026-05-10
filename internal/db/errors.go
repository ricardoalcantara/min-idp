package db

import "errors"

var ErrEntityNotFound = errors.New("not found")
var ErrEntityConflict = errors.New("conflict")
