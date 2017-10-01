package hatpear_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bluekeyes/hatpear"
)

func Example() {
	std := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("this failed!")
		hatpear.Store(r, err)
	})

	pear := hatpear.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("this also failed!")
	})

	mux := http.NewServeMux()
	mux.Handle("/std", std)
	mux.Handle("/hatpear", hatpear.Try(pear))

	catch := hatpear.Catch(func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	})

	http.ListenAndServe(":8000", catch(mux))
}
