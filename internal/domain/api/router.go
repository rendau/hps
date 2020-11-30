package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(a.hRoot).Methods("GET")

	return r
}
