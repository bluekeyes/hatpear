package hatpear_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bluekeyes/hatpear"
)

func TestStoreGet(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("test error")
		hatpear.Store(r, err)

		if stored := hatpear.Get(r); stored != err {
			t.Errorf("Stored error (%v) does not equal expected error (%v)", stored, err)
		}
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	catch := hatpear.Catch(func(w http.ResponseWriter, r *http.Request, err error) {})
	catch(h).ServeHTTP(w, r)
}

func TestTryCatch(t *testing.T) {
	var called bool
	var handledErr error

	catch := hatpear.Catch(func(w http.ResponseWriter, r *http.Request, err error) {
		called = true
		handledErr = err
	})

	err := errors.New("test error")
	h := hatpear.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return err
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	catch(hatpear.Try(h)).ServeHTTP(w, r)

	if !called {
		t.Error("Error handler function was not called by Catch()")
	}

	if handledErr != err {
		t.Errorf("Caught error (%v) does not equal expected error (%v)", handledErr, err)
	}
}

func TestStore_unconfigured(t *testing.T) {
	defer func() {
		if v := recover(); v == nil {
			t.Error("Store() with unconfigured request did not panic")
		}
	}()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	hatpear.Store(r, errors.New("test error"))
}

func TestGet_unconfigured(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := hatpear.Get(r); err != nil {
		t.Errorf("Get() with unconfigured request returned %v instead of nil", err)
	}
}

func TestRecover(t *testing.T) {
	var handledErr error

	rec := hatpear.Recover()
	catch := hatpear.Catch(func(w http.ResponseWriter, r *http.Request, err error) {
		handledErr = err
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	catch(rec(h)).ServeHTTP(w, r)

	if perr, ok := handledErr.(hatpear.PanicError); ok {
		if !strings.HasPrefix(perr.Error(), "panic:") {
			t.Errorf("Error string does not start with \"panic:\": %s", perr.Error())
		}

		v := perr.Value()
		if v != "test panic" {
			t.Errorf("Panic value (%v [%T]) does not equal expected value (test panic [string])", v, v)
		}

		if len(perr.StackTrace()) == 0 {
			t.Error("The stack trace associated with the panic error is empty")
		}

		hfunc := "hatpear_test.TestRecover.func2"
		trace := fmt.Sprintf("%+v", perr)

		if !strings.Contains(trace, hfunc) {
			t.Errorf("The stack trace does not contain the handler function (%s):\n%s", hfunc, trace)
		}
	} else {
		t.Errorf("Handled error with type \"%T\" was not a hatpear.PanicError", handledErr)
	}
}
