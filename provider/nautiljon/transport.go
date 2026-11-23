package nautiljon

import "net/http"

type rewriteTransport struct {
	base string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// no-op; real impl would use cloudscraper. kept for parity.
	return http.DefaultTransport.RoundTrip(req)
}
