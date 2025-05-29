package transport

import "net/http"

func AttachHostPrefix(schema, host string, next http.RoundTripper) http.RoundTripper {
	return Transporter(func(req *http.Request) (*http.Response, error) {
		if schema != "" {
			req.URL.Scheme = schema
		}
		if host != "" {
			req.Host = host
			req.URL.Host = host
		}

		return next.RoundTrip(req)
	})
}
