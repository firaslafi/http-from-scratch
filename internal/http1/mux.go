package http1

import (
	"net/http"
	"strings"
)

type urlPattern string
type httpMethod string

type router struct {
	routes map[urlPattern]routeRules
}

type routeRules struct {
	methods map[httpMethod]http.Handler
}

func New() *router {
	return &router{routes: make(map[urlPattern]routeRules)}
}

func (r *router) HandleFunc(method httpMethod, pattern urlPattern, f func(w http.ResponseWriter, req *http.Request)) {
	rules, exists := r.routes[pattern]
	if !exists {
		rules = routeRules{methods: make(map[httpMethod]http.Handler)}
		r.routes[pattern] = rules
	}
	rules.methods[method] = http.HandlerFunc(f)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// first we will try to find a registered URL pattern
	foundPattern, exists := r.routes[urlPattern(req.URL.Path)]
	if !exists {
		http.NotFound(w, req)
		return
	}
	// next we will try to check if such HTTP method was registered
	handler, exists := foundPattern.methods[httpMethod(req.Method)]
	if !exists {
		notAllowed(w, req, foundPattern)
		return
	}
	// finally we will call registered handler
	handler.ServeHTTP(w, req)
}

// small helper method
func notAllowed(w http.ResponseWriter, req *http.Request, r routeRules) {
	methods := make([]string, 0, len(r.methods))
	for m := range r.methods {
		methods = append(methods, string(m))
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello"))
}
