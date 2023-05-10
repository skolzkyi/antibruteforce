package internalhttp

import (
	"net/http"
)

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", loggingMiddleware(s.helloWorld, s.logg))
	mux.HandleFunc("/request/", loggingMiddleware(s.AuthorizationRequest, s.logg))
	mux.HandleFunc("/clearbucketbylogin/", loggingMiddleware(s.ClearBucketByLogin, s.logg))
	mux.HandleFunc("/clearbucketbyip/", loggingMiddleware(s.ClearBucketByIP, s.logg))
	mux.HandleFunc("/whitelist/", loggingMiddleware(s.WhiteList_REST, s.logg))
	mux.HandleFunc("/blacklist/", loggingMiddleware(s.BlackList_REST, s.logg))

	return mux
}
