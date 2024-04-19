# httpd

[![tests](https://github.com/shayanderson/httpd/actions/workflows/tests.yml/badge.svg)](https://github.com/shayanderson/httpd/actions/workflows/tests.yml)

Package `shayanderson/httpd` is a lightweight, fast HTTP router for Go. The router offers enhancements to the standard library `net/http` router.

## Features

- lightweight, simple and fast
- zero external dependencies
- `net/http` compatible
- middleware support
- named route parameters
- centralized error handling

## Requirements

- Go 1.22+

## Installation

```bash
go get github.com/shayanderson/httpd
```

## Example

```go
import (
    // ...
    "github.com/shayanderson/httpd"
)

func main() {
    r := httpd.New()

    // add middleware stack
    r.Use(httpd.LoggerMiddleware)
    r.Use(httpd.RecoverMiddleware)

    // add routes
    r.Get("/", indexRoute)
    r.Get("/api/user/{id}", apiUserRoute)

    // start server
    err := http.ListenAndServe(":8080", r)

    if err != nil {
        slog.Error("httpd server error", "err", err)
        os.Exit(1)
    }
}

// index route handler
func indexRoute(w http.ResponseWriter, r *http.Request) error {
    httpd.Respond(w, http.StatusOK, []byte("index"))
    return nil
}

// user route handler
func apiUserRoute(w http.ResponseWriter, r *http.Request) error {
    userID := r.PathValue("id")
    user, err := getUser(userID) // fetch user from DB or cache

    if err != nil {
        return err // return error
    }

    if user == nil {
        // return error with custom status code
        //   the `true` flag indicates it is safe to return the error message
        //   in the response like `{"error": "user not found"}`
        return httpd.NewError(http.StatusNotFound, errors.New("user not found"), true)
    }

    httpd.RespondJSON(w, http.StatusOK, map[string]string{
        "id": user.ID,
        "name": user.Name,
    })

    return nil
}

```

## Custom Error Handler

A custom error handler can be set using a function with the signature `func(http.ResponseWriter, *http.Request, error)`. Set a custom error handler example:

```go
func customErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
    // ...
}

func main() {
    httpd.DefaultErrorHandler = customErrorHandler
    r := httpd.New()
    // ...
}
```
