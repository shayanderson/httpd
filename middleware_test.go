package httpd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		status := 0
		rw := responseWriter{
			w:      &w,
			status: &status,
		}

		if rw.Header() == nil {
			t.Error("expected non-nil header")
		}

		if _, err := rw.Write([]byte("test")); err != nil {
			t.Error(err)
		}

		rw.WriteHeader(200)

		if *rw.status != 200 {
			t.Errorf("expected status 200, got %d", rw.status)
		}

		return errors.New("test error")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestLoggerMiddleware(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()
	r.Use(LoggerMiddleware)

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}

func TestRecoverMiddleware(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()
	r.Use(RecoverMiddleware)

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
}
