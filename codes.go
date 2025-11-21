package main

import (
	"errors"
)

var ErrNotFound = errors.New("This content could not be found")
var ErrOffline = errors.New("Player is offline")

func Error(message string) map[string]string {
	return map[string]string{
		"error": message,
	}
}
