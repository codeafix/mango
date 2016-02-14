package mango

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockRoutes struct {
	routes map[string]map[string]ContextHandlerFunc
}

func (m *mockRoutes) AddHandlerFunc(pattern, method string, handlerFunc ContextHandlerFunc) {
	_, ok := m.routes[pattern]
	if !ok {
		m.routes[pattern] = make(map[string]ContextHandlerFunc)
	}
	m.routes[pattern][method] = handlerFunc
}

func (m *mockRoutes) HandlerFuncs(path string) (map[string]ContextHandlerFunc, map[string]string, bool) {
	hm, ok := m.routes[path]
	return hm, nil, ok
}

func newMockRoutes() *mockRoutes {
	mr := mockRoutes{}
	mr.routes = make(map[string]map[string]ContextHandlerFunc)
	return &mr
}

func TestGetAddsHandlerToRoutes(t *testing.T) {
	want := "testFunc"
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Get("/test", testFunc)
	handlers, _, ok := rtr.routes.HandlerFuncs("/test")
	if !ok {
		t.Errorf("Handler not added")
	}
	h := handlers["GET"]
	got := extractFnName(h)
	if got != want {
		t.Errorf("Handler function = %q, want %q", got, want)
	}
}

func TestPostAddsHandlerToRoutes(t *testing.T) {
	want := "testFunc"
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Post("/test", testFunc)
	handlers, _, ok := rtr.routes.HandlerFuncs("/test")
	if !ok {
		t.Errorf("Handler not added")
	}
	h := handlers["POST"]
	got := extractFnName(h)
	if got != want {
		t.Errorf("Handler function = %q, want %q", got, want)
	}
}

func TestPutAddsHandlerToRoutes(t *testing.T) {
	want := "testFunc"
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Put("/test", testFunc)
	handlers, _, ok := rtr.routes.HandlerFuncs("/test")
	if !ok {
		t.Errorf("Handler not added")
	}
	h := handlers["PUT"]
	got := extractFnName(h)
	if got != want {
		t.Errorf("Handler function = %q, want %q", got, want)
	}
}

func TestPatchAddsHandlerToRoutes(t *testing.T) {
	want := "testFunc"
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Patch("/test", testFunc)
	handlers, _, ok := rtr.routes.HandlerFuncs("/test")
	if !ok {
		t.Errorf("Handler not added")
	}
	h := handlers["PATCH"]
	got := extractFnName(h)
	if got != want {
		t.Errorf("Handler function = %q, want %q", got, want)
	}
}

func TestDeleteAddsHandlerToRoutes(t *testing.T) {
	want := "testFunc"
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Del("/test", testFunc)
	handlers, _, ok := rtr.routes.HandlerFuncs("/test")
	if !ok {
		t.Errorf("Handler not added")
	}
	h := handlers["DELETE"]
	got := extractFnName(h)
	if got != want {
		t.Errorf("Handler function = %q, want %q", got, want)
	}
}

func TestSendErrorUsesSuppliedStatusCode(t *testing.T) {
	r := Router{}
	w := httptest.NewRecorder()
	r.sendError(w, "an error string", 404)
	want := 404
	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestSendErrorUsesSuppliedErrorMessage(t *testing.T) {
	r := Router{}
	w := httptest.NewRecorder()
	r.sendError(w, "an error string", 404)
	want := "an error string"
	got := w.Body.String()
	if got != want {
		t.Errorf("Body = %q, want %q", got, want)
	}
}

func TestSendErrorSetsContentTypeToTextPlain(t *testing.T) {
	r := Router{}
	w := httptest.NewRecorder()
	r.sendError(w, "an error string", 404)
	want := "text/plain; charset=utf-8"
	got := w.HeaderMap.Get("Content-Type")
	if got != want {
		t.Errorf("Body = %q, want %q", got, want)
	}
}

func TestWhenNoMatchingRouteServeHTTPReturns404NotFound(t *testing.T) {
	want := 404
	rtr := Router{}
	rtr.routes = newMockRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestWhenNoMatchingHandlerServeHTTPReturns405MethodNotAllowed(t *testing.T) {
	want := 405
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Del("/test", testFunc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestWhenNoErrorAndNoStatusSetServeHTTPReturns200OK(t *testing.T) {
	want := 200
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.Get("/test", testFunc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestAddPreHookAppendsToHookList(t *testing.T) {
	want := "prehookhandler"
	callStack := ""
	ph := func(ctx *Context) error {
		callStack += "prehook"
		return nil
	}

	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.AddPreHook(ph)
	rtr.Get("/test", func(ctx *Context) {
		callStack += "handler"
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := callStack
	if got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestPreHooksCalledInOrder(t *testing.T) {
	want := "prehook1prehook2prehook3handler"
	callStack := ""
	ph1 := func(ctx *Context) error {
		callStack += "prehook1"
		return nil
	}
	ph2 := func(ctx *Context) error {
		callStack += "prehook2"
		return nil
	}
	ph3 := func(ctx *Context) error {
		callStack += "prehook3"
		return nil
	}

	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.AddPreHook(ph1)
	rtr.AddPreHook(ph2)
	rtr.AddPreHook(ph3)
	rtr.Get("/test", func(ctx *Context) {
		callStack += "handler"
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := callStack
	if got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestAddPostHookAppendsToHookList(t *testing.T) {
	want := "handlerposthook"
	callStack := ""
	ph := func(ctx *Context) error {
		callStack += "posthook"
		return nil
	}

	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.AddPostHook(ph)
	rtr.Get("/test", func(ctx *Context) {
		callStack += "handler"
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := callStack
	if got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestPostHooksCalledInOrder(t *testing.T) {
	want := "handlerposthook1posthook2posthook3"
	callStack := ""
	ph1 := func(ctx *Context) error {
		callStack += "posthook1"
		return nil
	}
	ph2 := func(ctx *Context) error {
		callStack += "posthook2"
		return nil
	}
	ph3 := func(ctx *Context) error {
		callStack += "posthook3"
		return nil
	}

	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.AddPostHook(ph1)
	rtr.AddPostHook(ph2)
	rtr.AddPostHook(ph3)
	rtr.Get("/test", func(ctx *Context) {
		callStack += "handler"
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	rtr.ServeHTTP(w, req)

	got := callStack
	if got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestGetSimpleTextResponse(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()

	rtr.Get("/test", func(ctx *Context) {
		ctx.RespondWith("We're just two lost souls swimming in a fish bowl")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	rtr.ServeHTTP(w, req)
	want := "We're just two lost souls swimming in a fish bowl"
	got := w.Body.String()
	if got != want {
		t.Errorf("Response = %q, want %q", got, want)
	}
}

func TestGetResponseStatus(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithStatus(404)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	rtr.ServeHTTP(w, req)
	want := 404
	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestGetEncodedResponse(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.encoderEngine = &mockEncoderEngine{}

	var model = struct {
		a string
		b string
		c int
	}{
		"mango", "biscuits", 34,
	}

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithModel(fmt.Sprint(model))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "test/test")
	rtr.ServeHTTP(w, req)

	want := fmt.Sprint(model)
	got := w.Body.String()
	if got != want {
		t.Errorf("Response = %q, want %q", got, want)
	}
}

func TestResponseCodeWhenRequestAcceptHeaderIsUnsupported(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.encoderEngine = &mockEncoderEngine{}

	var model = struct {
		a string
		b string
		c int
	}{
		"mango", "biscuits", 34,
	}

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithModel(fmt.Sprint(model))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "test/mango")
	rtr.ServeHTTP(w, req)

	want := 406
	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestResponseMessageWhenRequestAcceptHeaderIsUnsupported(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.encoderEngine = &mockEncoderEngine{}

	var model = struct {
		a string
		b string
		c int
	}{
		"mango", "biscuits", 34,
	}

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithModel(fmt.Sprint(model))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "test/mango")
	rtr.ServeHTTP(w, req)

	want := "Unable to encode to requested acceptable formats: \"test/mango\""
	got := w.Body.String()
	if got != want {
		t.Errorf("Error message = %q, want %q", got, want)
	}
}

func TestResponseCodeWhenErrorEncodingPayload(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.encoderEngine = &mockEncoderEngine{}

	var model = struct {
		a string
		b string
		c int
	}{
		"mango", "biscuits", 34,
	}

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithModel(model)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "test/test")
	rtr.ServeHTTP(w, req)

	want := 500
	got := w.Code
	if got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
}

func TestResponseMessageWhenErrorEncodingPayload(t *testing.T) {
	rtr := Router{}
	rtr.routes = newMockRoutes()
	rtr.encoderEngine = &mockEncoderEngine{}

	var model = struct {
		a string
		b string
		c int
	}{
		"mango", "biscuits", 34,
	}

	rtr.Get("/test", func(ctx *Context) {
		ctx.Respond().WithModel(model)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "test/test")
	rtr.ServeHTTP(w, req)

	want := "Sorry, something went wrong."
	got := w.Body.String()
	if got != want {
		t.Errorf("Error message = %q, want %q", got, want)
	}
}

func TestNewRouterSetsRoutes(t *testing.T) {
	want := reflect.TypeOf(&tree{}).String()
	r := NewRouter()
	if r.routes == nil {
		t.Errorf("Routes type = \"<nil>\", want %q", want)
		return
	}
	got := reflect.TypeOf(r.routes).String()
	if got != want {
		t.Errorf("Routes type = %q, want %q", got, want)
	}
}

func TestNewRouterSetsEncoderEngine(t *testing.T) {
	want := reflect.TypeOf(&encoderEngine{}).String()
	r := NewRouter()
	if r.encoderEngine == nil {
		t.Errorf("EncoderEngine type = \"<nil>\", want %q", want)
		return
	}
	got := reflect.TypeOf(r.encoderEngine).String()
	if got != want {
		t.Errorf("EncoderEngine = %q, want %q", got, want)
	}
}

func TestNewRouterInitialisesEncoderEngineWithDefaultMediaType(t *testing.T) {
	want := DefaultMediaType
	r := NewRouter()
	got := r.encoderEngine.DefaultMediaType()
	if got != want {
		t.Errorf("EncoderEngine.DefaultMediaType = %q, want %q", got, want)
	}
}