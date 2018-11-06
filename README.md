# Fork from https://gist.github.com/reagent/043da4661d2984e9ecb1ccb5343bf438

# Custom HTTP Routing in Go

## Basic Routing

Responding to requests via simple route matching is built in to Go's [`net/http`](https://golang.org/pkg/net/http/) standard library package. Just register the path prefixes and callbacks you want invoked and then call the [`ListenAndServe`](https://golang.org/pkg/net/http/#ListenAndServe) to have the default request handler invoked on each request.  For example:

```go
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "Hello world\n")
	})

	err := http.ListenAndServe(":9000", nil)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
```

While it may look strange to pass `nil` as the second parameter to `ListenAndServe`, this causes Go to use the `DefaultServeMux` request multiplexer.  Requests to the configured endpoint look like this:

```
$ curl -i http://localhost:9000/hello
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Mon, 26 Jun 2017 23:58:40 GMT
Content-Length: 12

Hello world
```

By default, this handler will respond with a 404 when it can't find a match:


```
$ curl -i http://localhost:9000/
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Mon, 26 Jun 2017 23:59:11 GMT
Content-Length: 19

404 page not found
```

As you can see above, we're not controlling the content of the response, so the `Content-Type` header reverts to Go's default.  If you would prefer not to rely on the "magic" of the default multiplexer, you can configure your own by creating a [`ServeMux` instance](https://golang.org/pkg/net/http/#ServeMux).

## Routing with ServeMux

The process of registering callbacks with this method is similar to the previous example, but in this case we call [`HandleFunc`](https://golang.org/pkg/net/http/#ServeMux.HandleFunc) on the multiplexer that we create:

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.Replace(r.URL.Path, "/hello/", "", 1)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, fmt.Sprintf("Hello %s\n", name))
	})

	handler.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "Hello world\n")
	})

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)

		io.WriteString(w, "Not found\n")
	})

	err := http.ListenAndServe(":9000", handler)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
```

You'll notice two changes here:

1. The route prefix for `/hello/` that will match against any subtree that matches this prefix
1. The custom handler for `/` that responds with a 404 status when the request was not matched by any previous pattern

Here are the responses for each:

```
$ curl -i http://localhost:9000/hello/Patrick
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Tue, 27 Jun 2017 04:44:30 GMT
Content-Length: 14

Hello Patrick
```

```
$ curl -i http://localhost:9000/asdf
HTTP/1.1 404 Not Found
Content-Type: text/plain
Date: Tue, 27 Jun 2017 04:44:36 GMT
Content-Length: 10

Not found
```

The duplication present in each handler is something that can be easily refactored by moving away from interacting with `http.ResponseWriter` directly, so we'll do that next.

## Customizing the HTTP Response

The common tasks of writing the `Content-Type` header, setting the HTTP status, and returning the body can be moved to a single function by embedding `http.ResponseWriter` in a new struct and moving the duplicate code to our new struct.  For example:

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Response struct {
	http.ResponseWriter
}

func (r *Response) Text(code int, body string) {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(code)

	io.WriteString(r, fmt.Sprintf("%s\n", body))
}

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.Replace(r.URL.Path, "/hello/", "", 1)

		resp := Response{w}
		resp.Text(http.StatusOK, fmt.Sprintf("Hello %s", name))
	})

	handler.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		resp := Response{w}
		resp.Text(http.StatusOK, "Hello world")
	})

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := Response{w}
		resp.Text(http.StatusNotFound, "Not found")
	})

	err := http.ListenAndServe(":9000", handler)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
```

The only difference between this and the previous example is that we've condensed the 3 lines needed to write a response down to a single call to `Response.Text()` to set the status and send the body to the client.  While assigning a local `Response` instance is tedious, it's necessary in this case due to the specific function signature required by `HandleFunc`.  To see how we might simplify this, we'll have to write our own HTTP handler.

## Writing a Custom Handler

Per the [documentation](https://golang.org/pkg/net/http/#Handler), a struct can be treated as a handler if it implements the `ServeHTTP()` method.  So, if we wanted to ditch the request multiplexer altogether, we could respond to requests directly from a custom handler defined in a new struct:

```go
package main

import (
	"io"
	"log"
	"net/http"
)

type App struct{}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "Hello world\n")
}

func main() {
	err := http.ListenAndServe(":9000", &App{})

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
```

Since our `App` struct responds to `ServeHTTP`, we can pass it directly to `ListenAndServe` and it will receive all HTTP requests.  In this simple example, all requests receive the same response regardless of the request path:

```
$ curl -i http://localhost:9000/ok/pal
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Tue, 27 Jun 2017 04:58:39 GMT
Content-Length: 12

Hello world
```

This isn't very interesting in itself, but it *does* now give us the ability to wrap each incoming request before passing it off to a custom handler function.

## Custom Regular Expression-Based Router

By using a custom handler, we now have control over when the requests are intercepted, we can pass our own request and response structs to the matched route.  Since we are no longer using a `http.ServeMux` instance, I've introduced a custom regular expression-based router:


```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

type Handler func(*Response, *Request)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type App struct {
	Routes       []Route
	DefaultRoute Handler
}

func NewApp() *App {
	app := &App{
		DefaultRoute: func(resp *Response, req *Request) {
			resp.Text(http.StatusNotFound, "Not found")
		},
	}

	return app
}

func (a *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}

	a.Routes = append(a.Routes, route)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{Request: r}
	resp := &Response{w}

	for _, rt := range a.Routes {
		if matches := rt.Pattern.FindStringSubmatch(r.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				req.Params = matches[1:]
			}

			rt.Handler(resp, req)
			return
		}
	}

	a.DefaultRoute(resp, req)
}

type Request struct {
	*http.Request
	Params []string
}

type Response struct {
	http.ResponseWriter
}

func (r *Response) Text(code int, body string) {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(code)

	io.WriteString(r, fmt.Sprintf("%s\n", body))
}

func main() {
	app := NewApp()

	app.Handle(`^/hello$`, func(resp *Response, req *Request) {
		resp.Text(http.StatusOK, "Hello world")
	})

	app.Handle(`/hello/([\w\._-]+)$`, func(resp *Response, req *Request) {
		resp.Text(http.StatusOK, fmt.Sprintf("Hello %s", req.Params[0]))
	})

	err := http.ListenAndServe(":9000", app)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
```

The `App` struct now has a collection of routes, each with a corresponding callback.  If the request path matches the configured pattern, that callback will be triggered.  Otherwise, the default route will be invoked and the server will respond with a 404 status:

```
$ curl -i http://localhost:9000/hello
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Tue, 27 Jun 2017 14:39:24 GMT
Content-Length: 12

Hello world
```

```
$ curl -i http://localhost:9000/hello/Patrick
HTTP/1.1 200 OK
Content-Type: text/plain
Date: Tue, 27 Jun 2017 14:39:28 GMT
Content-Length: 14

Hello Patrick
```
```
$ curl -i http://localhost:9000/missing
HTTP/1.1 404 Not Found
Content-Type: text/plain
Date: Tue, 27 Jun 2017 14:39:32 GMT
Content-Length: 10

Not found
```

The custom `Request` struct is worth examining -- rather than performing substring replacement to determine the dynamic message, it keeps track of any capture groups in the route pattern and exposes them through the `Params` field.  While this gets the job done, it's not ideal from a design perspective.

## Wrapping it All in a Context

Rather than storing `Params` on the `Request` struct, we can instead introduce another struct to capture these values and use type embedding to have `http.Request` and`http.ResponseWriter` handle the traditional HTTP interactions.  This will also simplify the `Handler` signature as it now only needs to take a single `Context` struct and can perform all the response handling needed:

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

type Handler func(*Context)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type App struct {
	Routes       []Route
	DefaultRoute Handler
}

func NewApp() *App {
	app := &App{
		DefaultRoute: func(ctx *Context) {
			ctx.Text(http.StatusNotFound, "Not found")
		},
	}

	return app
}

func (a *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}

	a.Routes = append(a.Routes, route)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{Request: r, ResponseWriter: w}

	for _, rt := range a.Routes {
		if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}

			rt.Handler(ctx)
			return
		}
	}

	a.DefaultRoute(ctx)
}

type Context struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

func (c *Context) Text(code int, body string) {
	c.ResponseWriter.Header().Set("Content-Type", "text/plain")
	c.WriteHeader(code)

	io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
}

func main() {
	app := NewApp()

	app.Handle(`^/hello$`, func(ctx *Context) {
		ctx.Text(http.StatusOK, "Hello world")
	})

	app.Handle(`/hello/([\w\._-]+)$`, func(ctx *Context) {
		ctx.Text(http.StatusOK, fmt.Sprintf("Hello %s", ctx.Params[0]))
	})

	err := http.ListenAndServe(":9000", app)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}

```

This is just the start of what's possible when customizing an HTTP response handler.  You can take this further by:

* Matching requests against a specific HTTP method (e.g. GET / POST) and having different handler for the different request types
* Matching against `Content-Type` to invoke a separate handler for different content requests (e.g. `text/html`, `application/json`, etc...)
* Adding additional response methods (e.g. `ctx.JSON`) to send a more appropriate response for the requested `Content-Type`
