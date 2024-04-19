package httpd

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type Middleware func(http.Handler) http.Handler

type responseWriter struct {
	w      *http.ResponseWriter
	status *int
}

func (r responseWriter) Header() http.Header {
	return (*r.w).Header()
}

func (r responseWriter) Write(b []byte) (int, error) {
	return (*r.w).Write(b)
}

func (r responseWriter) WriteHeader(status int) {
	(*r.status) = status
	(*r.w).WriteHeader(status)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := 0
		rw := responseWriter{
			w:      &w,
			status: &status,
		}

		defer func() {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			slog.Info(
				fmt.Sprintf(
					"[httpd] %s %s://%s%s %s",
					r.Method,
					scheme,
					r.Host,
					r.RequestURI,
					r.Proto,
				),
				"from", r.RemoteAddr,
				"status", *rw.status,
				"took", time.Since(start),
			)
		}()

		next.ServeHTTP(rw, r)
	}

	return http.HandlerFunc(fn)
}

func RecoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				slog.Error("[httpd] recovering from panic", "err", err, "trace", debug.Stack())
				DefaultErrorHandler(w, r, fmt.Errorf("recovering from panic"))
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
