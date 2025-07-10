package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/OvsyannikovAlexandr/marketplace/api-service/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/users/login").Handler(proxyTo("http://user-service:8080"))
	r.PathPrefix("/users/register").Handler(proxyTo("http://user-service:8080"))

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTMiddleware)

	protected.PathPrefix("/users").Handler(proxyTo("http://user-service:8080"))
	protected.PathPrefix("/products").Handler(proxyTo("http://product-service:8080"))
	protected.PathPrefix("/cart").Handler(proxyTo("http://cart-service:8080"))
	protected.PathPrefix("/orders").Handler(proxyTo("http://order-service:8080"))

	return r
}

func proxyTo(target string) http.Handler {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if userID, ok := req.Context().Value("user_id").(int64); ok {
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
		}
	}

	return proxy
}
