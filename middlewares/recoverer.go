package middlewares

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
)

type MyResponseWriter struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (mrw *MyResponseWriter) WriteHeader(statusCode int) {
	mrw.status = statusCode
}

func (mrw *MyResponseWriter) Write(b []byte) (int, error) {
	return mrw.buf.Write(b)
}

func (mrw *MyResponseWriter) flush() {
	status := mrw.status
	if status == 0 {
		status = http.StatusOK
	}
	mrw.ResponseWriter.WriteHeader(status)
	mrw.ResponseWriter.Write(mrw.buf.Bytes())
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
			stackTrace := string(debug.Stack())
			log.Printf("stack trace: %s\n", stackTrace)

			message := r.message
			targetPattern := `(?m:^\s(.+):(\d+))`
			re := regexp.MustCompile(targetPattern)
			matches := re.FindAllStringSubmatch(stackTrace, -1)
			if matches != nil {
				stackTrace = re.ReplaceAllString(stackTrace, `<a href="/debug/?path=$1&line=$2">$0</a>`)
			}

			// if env is dev, send error and stack trace to client
			if r.dev {
				message = fmt.Sprintf("<h1>Uh oh!</h1><h3>Error</h3>%s\n\n<h3>Stack trace</h3><pre>%s</pre>", e, stackTrace)
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(r.statusCode)
			w.Write([]byte(message))

		}
	}()
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
