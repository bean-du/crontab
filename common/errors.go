package common

import "errors"

var (
	ErrLockOccupied = errors.New("lock was occupied")
)
