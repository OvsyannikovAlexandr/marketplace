package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
)

func NewRouter() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/users").Handler(proxyTo("http://user-service:8080"))
	r.PathPrefix("/products").Handler(proxyTo("http://product-service:8080"))
	r.PathPrefix("/cart").Handler(proxyTo("http://cart-service:8080"))
	r.PathPrefix("/orders").Handler(proxyTo("http://order-service:8080"))

	return r
}

func proxyTo(target string) http.Handler {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}
