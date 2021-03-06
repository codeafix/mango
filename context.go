package mango

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Response is an object used to facilitate building a response.
type Response struct {
	context *Context
	model   interface{}
	status  int
}

// WithModel sets the Model that will be serialized for the response.
// The serialization mechanism will depend on the request Accept header
// and the encoding DefaultMediaType.
// This method returns the Response object and can be chained.
func (r *Response) WithModel(m interface{}) *Response {
	r.context.model = m
	r.context.responseReady = true
	return r
}

// WithStatus sets the HTTP status code of the response.
// This method returns the Response object and can be chained.
func (r *Response) WithStatus(s int) *Response {
	r.context.status = s
	r.context.responseReady = true
	return r
}

// WithHeader adds a header to the response.
// This method returns the Response object and can be chained.
func (r *Response) WithHeader(key, value string) *Response {
	r.context.Writer.Header().Add(key, value)
	return r
}

// WithContentType sets the Content-Type header of the response.
// This method returns the Response object and can be chained.
func (r *Response) WithContentType(ct string) *Response {
	r.context.Writer.Header().Set("Content-Type", ct)
	return r
}

// Context is the request context.
// Context encapsulates the underlying Req and Writer, but exposes
// them if required. It provides many helper methods which are
// designed to keep your handler code clean and free from boiler
// code.
type Context struct {
	Request       *http.Request
	Writer        http.ResponseWriter
	status        int
	payload       []byte
	model         interface{}
	RouteParams   map[string]string
	encoderEngine EncoderEngine
	Reader        io.ReadCloser
	Identity      Identity
	responseReady bool
}

// ContextHandlerFunc type is an adapter to allow the use of ordinary
// functions as HTTP handlers. It is similar to the standard library's
// http.HandlerFunc in that if f is a function with the appropriate
// signature, ContextHandlerFunc(f) is a Handler object that calls f.
type ContextHandlerFunc func(*Context)

// ServeHTTP calls f(c).
func (f ContextHandlerFunc) ServeHTTP(c *Context) {
	f(c)
}

// Respond returns a new context based Response object.
func (c *Context) Respond() *Response {
	return &Response{context: c}
}

// RespondWith is a generic method for producing a simple response.
// It takes a single parameter whose type will determine the action.
//
// Strings will be used for the response content.
// Integers will be used for the response status code.
// Any other type is deemed to be a model which will be serialized.
//
// The serialization mechanism for the model will depend on the
// request Accept header and the encoding DefaultMediaType.
// This method returns the Response object and can be chained.
func (c *Context) RespondWith(d interface{}) *Response {
	response := &Response{context: c}

	switch t := d.(type) {
	case int:
		c.status = t
	case string:
		c.payload = []byte(t)
	default: //must be a model
		c.model = d
	}
	c.responseReady = true
	return response
}

// Authenticated returns true if a request user has been authenticated.
// Authentication should be performed in a pre-hook, assigning a valid
// Identity to the Context if authentication succeeds.
// This method simply examines whether the Context has a valid Identity.
func (c *Context) Authenticated() bool {
	return c.Identity != nil
}

func (c *Context) urlSchemeHost() string {
	if c.Request.TLS != nil {
		return "https://" + c.Request.Host
	}
	return "http://" + c.Request.Host
}

// Error sends the specified message and HTTP status code as a response.
// Request handlers should cease execution after calling this method.
func (c *Context) Error(msg string, code int) {
	http.Error(c.Writer, msg, code)
}

// Redirect sends a redirect response using the specified URL and HTTP
// status.
// Request handlers should cease execution after calling this method.
// TODO: Not yet implemented
func (c *Context) Redirect(urlStr string, code int) {
	http.Redirect(c.Writer, c.Request, urlStr, code)
}

// // Render executes a template using the supplied data.
// // Request handlers should cease execution after calling this method.
// // TODO: Not yet implemented
// func (c *Context) Render(tmpl string, data interface{}) {
// 	panic("not yet implemented")
// }

//
// func (c *Context) sendResponse() {
// 	fmt.Fprintf(c.Writer, "")
// }

func (c *Context) contentDecoder() (Decoder, error) {
	ct := c.Request.Header.Get("Content-Type")
	return c.encoderEngine.GetDecoder(c.Request.Body, ct)
}

func (c *Context) acceptableMediaTypes() []string {
	hdr := c.Request.Header.Get("Accept")
	hdr = strings.Replace(hdr, " ", "", -1)
	types := strings.Split(hdr, ",")
	mt := make(mediaTypes, len(types))

	for i, t := range types {
		m, err := newMediaType(t)
		if err != nil {
			continue
		}
		mt[i] = *m
	}
	sort.Sort(mt)
	r := []string{}
	for _, t := range mt {
		r = append(r, t.String())
	}
	return r
}

// GetEncoder returns an Encoder suitable for serializing data in a response.
// The Encoder is selected based on the request Accept header (or default media
// type if no Accept header supplied).
// If successful, the an encoder and content-type are returned and a nil error.
// Success is determined by a nil error.
// The returned encoder will have been pre-injected with an io.Writer, so the
// Encode method can be called directly, passing the data to be encoded as the
// only parameter.
func (c *Context) GetEncoder() (Encoder, string, error) {
	mts := c.acceptableMediaTypes()
	var err error
	var mt string
	for _, mt = range mts {
		if mt == "*/*" {
			mt = c.encoderEngine.DefaultMediaType()
		}
		var encoder Encoder
		encoder, err = c.encoderEngine.GetEncoder(c.Writer, mt)
		if err == nil {
			return encoder, mt, nil
		}
	}
	return nil, mt, err
}

// Bind populates the supplied model with data from the request.
// This is performed in stages. initially, any requestbody content is
// deserialized.
//
// TODO: Following is not yet implemented:
//
// Route parameters are used next to populate any unset members.
// Finally, query parameters are used to populate any remaining unset members.
//
// This method is under review - currently Binding only uses deserialized
// request body content.
func (c *Context) Bind(m interface{}) error {
	decoder, err := c.contentDecoder()
	if err != nil {
		return fmt.Errorf("unable to bind: %v", err)
	}
	err = decoder.Decode(m)
	if err != nil {
		return fmt.Errorf("unable to bind: %v", err)
	}

	// TODO: now update any missing empty properties from url path/query params

	return nil
}
