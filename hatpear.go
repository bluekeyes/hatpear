package hatpear

import (
	"context"
	"errors"
	"net/http"
)

type contextKey int

const (
	errorKey contextKey = iota
)

// Store stores an error into the request's context. It panics if the request
// was not configured by the middleware to store errors.
func Store(r *http.Request, err error) {
	errptr, ok := r.Context().Value(errorKey).(*error)
	if !ok {
		panic("hatpear: request not configured to store errors")
	}
	*errptr = err
}

// Get retrieves an error from the request's context.
func Get(r *http.Request) error {
	errptr, ok := r.Context().Value(errorKey).(*error)
	if !ok {
		return errors.New("hatpear: request not configured to store errors")
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
