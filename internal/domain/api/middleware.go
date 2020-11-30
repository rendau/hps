package api

import (
	"github.com/rs/cors"
	"net/http"
)

func (a *St) middleware(h http.Handler) http.Handler {
	h = cors.AllowAll().Handler(h)

	return h
}
