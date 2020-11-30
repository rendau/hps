package api

import (
	"net/http"
)

func (a *St) hRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
