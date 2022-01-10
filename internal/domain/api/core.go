package api

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (a *St) sendTargetRequest(ctx context.Context, method, uri string, params url.Values, headers http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	for k, hv := range headers {
		for _, v := range hv {
			req.Header.Add(k, v)
		}
	}

	client := http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return client.Do(req)
}

func (a *St) retrieveRemoteIP(r *http.Request) (result string) {
	result = ""

	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		result = parts[0]
	}

	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			result = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			result = ip.String()
		}
	}

	return
}
