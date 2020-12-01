package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(a.hRoot)

	return a.middleware(r)
}
