package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/CodeGophercises/recoverer/middlewares"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	mux.HandleFunc("/debug/", sourceCodeHandler)
}

func main() {
	mux := http.NewServeMux()
	registerHandlers(mux)
	// wrap the mux with recoverer
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	isDev := env == "dev"
	recovererMux := middlewares.NewRecoverer(mux, isDev)
	log.Fatal(http.ListenAndServe(":3000", recovererMux))
}

func sourceCodeHandler(w http.ResponseWriter, req *http.Request) {
	// Asssumes existence of two query params: line and path
	values := req.URL.Query()
	if _, ok := values["line"]; !ok {
		http.Error(w, "missing line query param", http.StatusNotFound)
		return
	}
	if _, ok := values["path"]; !ok {
		http.Error(w, "missing path query param", http.StatusNotFound)
		return
	}
	line, path := values["line"][0], values["path"][0]
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	code, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	lexer := lexers.Get("go")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, err := lexer.Tokenise(nil, string(code))
	if err != nil {
		panic(err)
	}
	lineNum, err := strconv.Atoi(line)
	if err != nil {
		panic(err)
	}
	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}
	w.Header().Set("Content-Type", "text/html")
	formatter := html.New(html.WithLineNumbers(true), html.HighlightLines([][2]int{{lineNum, lineNum}}))
	formatter.Format(w, style, iterator)
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
