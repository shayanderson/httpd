package httpd

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

var DefaultErrorHandler ErrorHandler = defaultErrorHandler

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	resErr := "internal server error"

	if e, ok := err.(Error); ok {
		status = e.Status()
		if e.IsResponseError() {
			resErr = e.Error()
		}
	}

	slog.Error("[httpd]", "err", err, "status", status)

	RespondJSON(
		w,
		status,
		map[string]string{"error": resErr},
	)
}

type Error interface {
	error
	IsResponseError() bool
	Status() int
}

type StatusError struct {
	err             error
	isResponseError bool
	status          int
}

func NewError(status int, err error, isResponseError ...bool) Error {
	return StatusError{
		err:             err,
		isResponseError: len(isResponseError) > 0 && isResponseError[0],
		status:          status,
	}
}

func (e StatusError) Error() string {
	return e.err.Error()
}

func (e StatusError) IsResponseError() bool {
	return e.isResponseError
}

func (e StatusError) Status() int {
	return e.status
}

type Route func(http.ResponseWriter, *http.Request) error

func (r Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := r(w, req); err != nil {
		DefaultErrorHandler(w, req, err)
	}
}

func Respond(w http.ResponseWriter, status int, payload []byte) {
	w.WriteHeader(status)
	w.Write(payload)
}

func RespondJSON(w http.ResponseWriter, status int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	Respond(w, status, body)

	return nil
}

type Router struct {
	mux *http.ServeMux
	mw  []Middleware
}

func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
		mw:  []Middleware{},
	}
}

func (r *Router) route(method string, pattern string, route Route, middleware ...Middleware) {
	var h http.Handler = route
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	r.mux.Handle(method+" "+pattern, h)
}

func (r *Router) Delete(pattern string, route Route, middleware ...Middleware) {
	r.route(http.MethodDelete, pattern, route, middleware...)
}

func (r *Router) Get(pattern string, route Route, middleware ...Middleware) {
	r.route(http.MethodGet, pattern, route, middleware...)
}

func (r *Router) Handle(method string, pattern string, route Route, middleware ...Middleware) {
	r.route(method, pattern, route, middleware...)
}

func (r *Router) Patch(pattern string, route Route, middleware ...Middleware) {
	r.route(http.MethodPatch, pattern, route, middleware...)
}

func (r *Router) Post(pattern string, route Route, middleware ...Middleware) {
	r.route(http.MethodPost, pattern, route, middleware...)
}

func (r *Router) Put(pattern string, route Route, middleware ...Middleware) {
	r.route(http.MethodPut, pattern, route, middleware...)
}

func (r *Router) Use(middleware ...Middleware) {
	r.mw = append(r.mw, middleware...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var h http.Handler = r.mux
	for i := len(r.mw) - 1; i >= 0; i-- {
		h = r.mw[i](h)
	}

	h.ServeHTTP(w, req)
}
