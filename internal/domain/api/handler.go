package api

import (
	"io"
	"net/http"
	"strings"
)

func (a *St) hRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqHeaders := r.Header

	reqHOrigin := reqHeaders.Get("Origin")

	if reqHeaders.Get("X-Real-IP") == "" &&
		reqHeaders.Get("X-Forwarded-For") == "" &&
		reqHeaders.Get("X-Forwarded-Proto") == "" {
		if clientAddr := a.retrieveRemoteIP(r); clientAddr != "" {
			reqHeaders.Set("X-Real-IP", clientAddr)
			reqHeaders.Set("X-Forwarded-For", clientAddr)
			reqHeaders.Set("X-Forwarded-Proto", "http")
		}
	}

	resp, err := a.sendTargetRequest(ctx, r.Method, a.targetUri+r.URL.Path, r.URL.Query(), reqHeaders, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("proxy: fail to request target URL, error: " + err.Error()))
		return
	}
	defer resp.Body.Close()

	for k, hv := range resp.Header {
		for _, v := range hv {
			if k != "Set-Cookie" {
				w.Header().Add(k, v)
			}
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", reqHOrigin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// originDomain := ""
	//
	// if originUrl, err := url.Parse(reqHOrigin); err == nil {
	// 	originDomain = originUrl.Hostname()
	// }

	domainFromHost := ""

	if r.Host != "" {
		if i := strings.LastIndexByte(r.Host, ':'); i > -1 {
			domainFromHost = r.Host[:i]
		}
	}

	for _, cookie := range resp.Cookies() {
		cookie.Domain = domainFromHost
		http.SetCookie(w, cookie)
	}

	w.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(w, resp.Body)
}
