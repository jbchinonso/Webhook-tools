// Package errors defines the layered error model used across the application.
//
// Design (FPiS influence):
//   - Errors are VALUES, not control flow. No panics.
//   - Each architectural layer wraps the layer below with its own context:
//         RepoError -> ServiceError -> HandlerError
//   - The full stack of context is preserved as the error propagates upward.
//   - Only the handler layer converts a domain error into an HTTP response.
package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// Kind classifies an error at the boundary where it will be interpreted.
// It is the ONLY field the HTTP layer consults to pick a status code.
type Kind uint8

const (
	KindInternal Kind = iota
	KindValidation
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
	KindUnavailable
	KindTooManyRequests
)

func (k Kind) String() string {
	switch k {
	case KindValidation:
		return "validation"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindUnavailable:
		return "unavailable"
	case KindTooManyRequests:
		return "too_many_requests"
	default:
		return "internal"
	}
}

// BaseError is the shared kernel of every layered error. It carries the
// machine-readable code, a human-readable message, the wrapped cause, and
// structured metadata for logs.
type BaseError struct {
	Kind     Kind
	Code     string
	Message  string
	Cause    error
	Metadata map[string]any
}

func New(kind Kind, code, message string) *BaseError {
	return &BaseError{Kind: kind, Code: code, Message: message}
}

func Wrap(cause error, kind Kind, code, message string) *BaseError {
	return &BaseError{Kind: kind, Code: code, Message: message, Cause: cause}
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BaseError) Unwrap() error { return e.Cause }

// With attaches structured metadata for logs. Chainable.
func (e *BaseError) With(key string, value any) *BaseError {
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	e.Metadata[key] = value
	return e
}

func (e *BaseError) base() *BaseError { return e }

// -----------------------------------------------------------------------------
// Repo layer
// -----------------------------------------------------------------------------

type RepoError struct {
	*BaseError
	Op         string
	Table      string
	Constraint string
}

func NewRepo(op, table string, cause error) *RepoError {
	return &RepoError{
		BaseError: Wrap(cause, KindInternal, "repo.error", "database operation failed"),
		Op:        op,
		Table:     table,
	}
}

func (e *RepoError) WithConstraint(c string) *RepoError {
	e.Constraint = c
	return e
}

// -----------------------------------------------------------------------------
// Service layer
// -----------------------------------------------------------------------------

type ServiceError struct {
	*BaseError
	Domain string
}

func NewService(domain string, kind Kind, code, message string, cause error) *ServiceError {
	var be *BaseError
	if cause != nil {
		be = Wrap(cause, kind, code, message)
	} else {
		be = New(kind, code, message)
	}
	return &ServiceError{BaseError: be, Domain: domain}
}

// -----------------------------------------------------------------------------
// Handler layer
// -----------------------------------------------------------------------------

type HandlerError struct {
	*BaseError
	Status int
}

// ToHandler walks the error chain looking for the outermost *BaseError.
// Unknown errors become a generic 500 without leaking internals.
func ToHandler(err error) *HandlerError {
	if be := findBase(err); be != nil {
		return &HandlerError{BaseError: be, Status: statusFromKind(be.Kind)}
	}
	return &HandlerError{
		BaseError: New(KindInternal, "internal", "internal server error"),
		Status:    http.StatusInternalServerError,
	}
}

func findBase(err error) *BaseError {
	for err != nil {
		if c, ok := err.(interface{ base() *BaseError }); ok {
			return c.base()
		}
		err = stderrors.Unwrap(err)
	}
	return nil
}

func statusFromKind(k Kind) int {
	switch k {
	case KindValidation:
		return http.StatusUnprocessableEntity
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func Is(err, target error) bool { return stderrors.Is(err, target) }
func As(err error, target any) bool { return stderrors.As(err, target) }

// IsLayered reports whether err (or anything it wraps) is one of our layered
// errors — BaseError, RepoError, ServiceError, or HandlerError. It works by
// walking the unwrap chain looking for the base() marker that all of them
// promote from *BaseError.
//
// Use this instead of errors.As when the error may have crossed a library
// boundary (e.g. a GORM transaction) that can re-wrap it.
func IsLayered(err error) bool {
	return findBase(err) != nil
}