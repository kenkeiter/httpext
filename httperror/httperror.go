/*
Package httperror allows convenient generation of consistently-structured
errors for HTTP APIs.
*/
package httperror

import (
	"fmt"
)

/*
Error defines an interface that fulfills the error interface, and adds support
for additional parameters necessary to provide well-structured errors in for
HTTP APIs.
*/
type Error interface {
	error

	// Status returns the HTTP status code of the error.
	Status() int

	// ID returns a developer-managed, service-unique, machine-readable
	// identifier for the error.
	ID() string

	// Message provides a human-readable string describing the error that
	// occurred in one or two sentences.
	Message() string

	// Detail may optionally contain a marshallable interface{} providing
	// additional contextual information about the error.
	Detail() interface{}

	// Equal compares one Error with another, returning true if the errors are
	// the same. Fields compared typically include ID, Status, and Message.
	Equal(Error) bool

	// Marshal provides a representative structure for the error, for use when
	// marshalling the error into a textual or machine-readable format.
	Marshal() (interface{}, error)

	// WithDetail clones the error, and creates a derivative instance that
	// includes the detail interface{} provided.
	WithDetail(interface{}) Error
}

type httpError struct {
	status  int
	id      string
	message string
	detail  interface{}
}

// New creates a new type of error, given an HTTP status code, unique
// identifying string, and message.
func New(status int, id, message string) Error {
	return &httpError{
		id:      id,
		status:  status,
		message: message,
	}
}

// Error provides a string representation, and conforms the httperror
// interface to Go's built-in error interface.
func (e *httpError) Error() string {
	if e.Detail() != nil {
		return fmt.Sprintf("%s (%v) <HTTP %d:%s>", e.Message(), e.Detail(), e.Status(), e.ID())
	}
	return fmt.Sprintf("%s <HTTP %d:%s>", e.Message(), e.Status(), e.ID())
}

// ID returns a unique identifying string associated with the error.
func (e *httpError) ID() string {
	return e.id
}

// Status returns the HTTP status code associated with the error.
func (e *httpError) Status() int {
	return e.status
}

// Message returns the message associated with the error.
func (e *httpError) Message() string {
	return e.message
}

// Detail returns the arbitrary detail interface{} associated with the error.
func (e *httpError) Detail() interface{} {
	return e.detail
}

// Equal compares the status code and message of two Errors to determine if
// they are identical. Equal does not compare detail, however; this is by
// design to make error types more generalizable.
func (e *httpError) Equal(err Error) bool {
	return e.ID() == err.ID() &&
		e.Status() == err.Status() &&
		e.Message() == err.Message()
}

// Marshal provides an arbitrary representation of the details of the error.
func (e *httpError) Marshal() (interface{}, error) {
	repr := struct {
		ID      string      `json:"id"`
		Message string      `json:"message"`
		Detail  interface{} `json:"detail,omitempty"`
	}{e.ID(), e.Message(), e.Detail()}
	return repr, nil
}

// WithDetail clones the Error and creates a new instance, setting the detail
// attribute.
func (e *httpError) WithDetail(detail interface{}) Error {
	derivedErr := e.clone()
	derivedErr.detail = detail
	return derivedErr
}

func (e *httpError) clone() *httpError {
	return &httpError{
		id:      e.id,
		status:  e.status,
		message: e.message,
		detail:  e.detail,
	}
}
