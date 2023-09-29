package webserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

type WebServer struct {
	Router        chi.Router
	Handlers      map[string]http.HandlerFunc
	WebServerPort string
}

func NewWebServer(webServerPort string) *WebServer {
	return &WebServer{
		WebServerPort: webServerPort,
		Handlers:      make(map[string]http.HandlerFunc),
		Router:        chi.NewRouter(),
	}
}

func (this *WebServer) AddHandler(path string, handler http.HandlerFunc) {
	this.Handlers[path] = handler
}

func (this *WebServer) Start() {
	this.Router.Use(middleware.Logger)
	for path, handler := range this.Handlers { //register handlers
		this.Router.HandleFunc(path, handler)
	}
	err := http.ListenAndServe(this.WebServerPort, this.Router)
	if err != nil {
		panic(err.Error())
	}
}
