package entities

import "errors"

var (
	ErrInternalServer    = errors.New("internal server error")
	ErrInventoryNotFound = errors.New("inventory not found")
)
