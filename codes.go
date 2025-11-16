package main

import "errors"

var ErrNotFound = errors.New("This content could not be found")
var ErrOffline = errors.New("Player is offline")
