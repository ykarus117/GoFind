package Store

import "fmt"

type InternalError struct {
	Message string
	Err     error
}
type FormatError struct {
	Message string
	Err     error
}
type NotFoundError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("Storage internal error from: %s \n %v", e.Message, e.Err)
}
func (e *FormatError) Error() string {
	return fmt.Sprintf("storage request format error: %s \n %v", e.Message, e.Err)
}
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Resource not found: %s \n %v", e.Message, e.Err)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}
func (e *FormatError) Unwrap() error {
	return e.Err
}
func (e *NotFoundError) Unwrap() error {
	return e.Err
}
