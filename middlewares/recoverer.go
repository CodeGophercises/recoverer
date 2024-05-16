package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// Interface embedded in a struct. Classic.
type MyResponseWriter struct {
	http.ResponseWriter
	status int
	data   []byte
}

// WriteHeader(statusCode int)
func (mrw *MyResponseWriter) WriteHeader(statusCode int) {
	// shadow the original impl
	mrw.status = statusCode
}

// Write([]byte) (int, error)
func (mrw *MyResponseWriter) Write(b []byte) (int, error) {
	// Dont actually write, that is irreversible
	mrw.data = append(mrw.data, b...)
	return len(b), nil
}

func (mrw *MyResponseWriter) flush() {
	status := mrw.status
	if status == 0 {
		status = http.StatusOK
	}
	mrw.ResponseWriter.WriteHeader(status)

	mrw.ResponseWriter.Write(mrw.data)
}

// Recoverer recovers the wrapped handler from panicking and sends out appropriate client response
type recoverer struct {
	// configurable options for client response
	statusCode int
	message    string
	handler    http.Handler
	dev        bool
}

func (r *recoverer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			// Logs the error and the stack trace
			log.Printf("got error:  %s\n", e)
			log.Printf("stack trace: %s\n", string(debug.Stack()))

			message := r.message

			// if env is dev, send error and stack trace to client
			if r.dev {
				message = fmt.Sprintf("Error:%s\n\n%s", e, string(debug.Stack()))
			}
			http.Error(w, message, r.statusCode)

		}
	}()

	// Wrap w in our own ResponseWriter that we can control
	nw := &MyResponseWriter{}
	nw.ResponseWriter = w
	r.handler.ServeHTTP(nw, req)
	nw.flush()
}

// Has sensible defaults
func NewRecoverer(h http.Handler, dev bool) *recoverer {
	return &recoverer{
		statusCode: http.StatusInternalServerError,
		message:    "Something went wrong",
		handler:    h,
		dev:        dev,
	}
}
