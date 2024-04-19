package httpd

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setup(t *testing.T) {
	t.Helper()
	log.SetOutput(io.Discard)
}

func teardown(t *testing.T) {
	t.Helper()
	log.SetOutput(os.Stdout)
}

func TestDefaultErrorHandler(t *testing.T) {
	setup(t)
	defer teardown(t)

	w := httptest.NewRecorder()
	err := errors.New("test error")

	DefaultErrorHandler(w, nil, err)

	if w.Code != 500 {
		t.Errorf("expected status code 500, got %d", w.Code)
	}

	expected := `{"error":"internal server error"}`
	if w.Body.String() != expected {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestDefaultErrorHandlerStatusError(t *testing.T) {
	setup(t)
	defer teardown(t)

	w := httptest.NewRecorder()
	err := NewError(404, errors.New("test error"), true)

	DefaultErrorHandler(w, nil, err)

	if w.Code != 404 {
		t.Errorf("expected status code 404, got %d", w.Code)
	}

	expected := `{"error":"test error"}`
	if w.Body.String() != expected {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestRouteError(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		return errors.New("test error")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code 500, got %d", w.Code)
	}

	expected := `{"error":"internal server error"}`
	if w.Body.String() != expected {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestRoutes(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Delete("/test/{id}", func(w http.ResponseWriter, req *http.Request) error {
		id := req.PathValue("id")
		return RespondJSON(w, http.StatusOK, map[string]string{"id": id})
	})

	r.Get("/test/{id}", func(w http.ResponseWriter, req *http.Request) error {
		id := req.PathValue("id")
		return RespondJSON(w, http.StatusOK, map[string]string{"id": id})
	})

	r.Patch("/test/{id}", func(w http.ResponseWriter, req *http.Request) error {
		id := req.PathValue("id")
		return RespondJSON(w, http.StatusOK, map[string]string{"id": id})
	})

	r.Post("/test/{id}", func(w http.ResponseWriter, req *http.Request) error {
		id := req.PathValue("id")
		return RespondJSON(w, http.StatusOK, map[string]string{"id": id})
	})

	r.Put("/test/{id}", func(w http.ResponseWriter, req *http.Request) error {
		id := req.PathValue("id")
		return RespondJSON(w, http.StatusOK, map[string]string{"id": id})
	})

	tests := []struct {
		method string
		path   string
	}{
		{"DELETE", "/test/123"},
		{"GET", "/test/123"},
		{"PATCH", "/test/123"},
		{"POST", "/test/123"},
		{"PUT", "/test/123"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			w := httptest.NewRecorder()

			req := httptest.NewRequest(test.method, test.path, nil)
			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("expected status code 200, got %d", res.StatusCode)
			}

			expected := `{"id":"123"}`
			data, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatal(err)
			}

			if string(data) != expected {
				t.Errorf("expected body %s, got %s", expected, string(data))
			}
		})
	}
}

func TestRouteMethod(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Handle(http.MethodHead, "/test", func(w http.ResponseWriter, req *http.Request) error {
		Respond(w, http.StatusOK, nil)
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}
}

func Test404(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code 404, got %d", w.Code)
	}
}

func Test405(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status code 405, got %d", w.Code)
	}
}

func TestMiddleware(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Test", "test")
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		Respond(w, http.StatusOK, nil)
		return nil
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	if w.Header().Get("X-Test") != "test" {
		t.Errorf("expected header X-Test=test, got %s", w.Header().Get("X-Test"))
	}
}

func TestRouteMiddleware(t *testing.T) {
	setup(t)
	defer teardown(t)

	r := New()

	r.Get("/test", func(w http.ResponseWriter, req *http.Request) error {
		Respond(w, http.StatusOK, nil)
		return nil
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Test", "test")
			next.ServeHTTP(w, req)
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	if w.Header().Get("X-Test") != "test" {
		t.Errorf("expected header X-Test=test, got %s", w.Header().Get("X-Test"))
	}
}
