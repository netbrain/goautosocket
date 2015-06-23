package main

// ----------------------------------------------------------------------------

// Error is the error type of the GAS package.
type Error int

const (
	// ErrMaxRetries is returned when the called function failed after the
	// maximum number of allowed tries.
	ErrMaxRetries = iota
)
