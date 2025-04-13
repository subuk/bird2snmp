package main

import "fmt"

// Error types for different categories of errors
type (
	// BirdError represents errors from BIRD daemon communication
	BirdError struct {
		Operation string
		Err       error
	}

	// SNMPError represents errors from SNMP operations
	SNMPError struct {
		Operation string
		Err       error
	}

	// ParseError represents errors from parsing BIRD output
	ParseError struct {
		Input string
		Err   error
	}
)

// Error implementations
func (e *BirdError) Error() string {
	return fmt.Sprintf("bird error during %s: %v", e.Operation, e.Err)
}

func (e *BirdError) Unwrap() error {
	return e.Err
}

func (e *SNMPError) Error() string {
	return fmt.Sprintf("snmp error during %s: %v", e.Operation, e.Err)
}

func (e *SNMPError) Unwrap() error {
	return e.Err
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error for input '%s': %v", e.Input, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Helper functions to create errors
func newBirdError(op string, err error) error {
	return &BirdError{Operation: op, Err: err}
}

func newSNMPError(op string, err error) error {
	return &SNMPError{Operation: op, Err: err}
}

func newParseError(input string, err error) error {
	return &ParseError{Input: input, Err: err}
}
