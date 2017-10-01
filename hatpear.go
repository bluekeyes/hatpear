package hatpear

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
)

type contextKey int

const (
	errorKey contextKey = iota
)

// Store stores an error into the request's context. It panics if the request
// was not configured to store errors.
func Store(r *http.Request, err error) {
	errptr, ok := r.Context().Value(errorKey).(*error)
	if !ok {
		panic("hatpear: request not configured to store errors")
	}
	*errptr = err
}

// Get retrieves an error from the request's context. It returns nil if the
// request was not configured to store errors.
func Get(r *http.Request) error {
	errptr, ok := r.Context().Value(errorKey).(*error)
	if !ok {
		return nil
	}
	return *errptr
}

// Middleware adds additional functionality to an existing handler.
type Middleware func(http.Handler) http.Handler

// Catch creates middleware that processes errors stored while serving a
// request. Errors are passed to the callback, which should write them to the
// response in an appropriate format.
func Catch(h func(w http.ResponseWriter, r *http.Request, err error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			ctx := context.WithValue(r.Context(), errorKey, &err)

			next.ServeHTTP(w, r.WithContext(ctx))
			if err != nil {
				h(w, r, err)
			}
		})
	}
}

// Handler is a variant on http.Handler that can return an error.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is a variant on http.HandlerFunc that can return an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// Try converts a handler to a standard http.Handler, storing any error in the
// request's context.
func Try(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Store(r, h.ServeHTTP(w, r))
	})
}

var (
	// RecoverStackSize is the max size of a stack for a recovered panic.
	RecoverStackSize = 4096
)

// Recover creates middleware that can recover from a panic in a handler,
// storing a *PanicError for future handling.
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					stack := make([]byte, RecoverStackSize)
					length := runtime.Stack(stack, false)

					Store(r, &PanicError{
						Value: v,
						Stack: stack[:length],
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// PanicError is an Error created from a recovered panic.
type PanicError struct {
	// Value is the value with which panic() was called
	Value interface{}
	// Stack is the stack trace of the panicing goroutine.
	Stack []byte
}

func (e *PanicError) Error() string {
	if err, ok := e.Value.(error); ok {
		return err.Error()
	}
	return fmt.Sprintf("%v", e.Value)
}
