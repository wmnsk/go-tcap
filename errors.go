package tcap

import "fmt"

// InvalidCodeError indicates that Code in TCAP message is invalid.
type InvalidCodeError struct {
	Code int
}

// Error returns error message with violating content.
func (e *InvalidCodeError) Error() string {
	return fmt.Sprintf("tcap: got invalid code: %d", e.Code)
}
