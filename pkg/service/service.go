package service

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/{id}", func(r chi.Router) {
		r.Post("/status", status)
		r.Get("/kubeconfig", kubeconfig)
		workDir, _ := os.Getwd()
		filesDir := filepath.Join(workDir, "dashboard")
		FileServer(r, "/dashboard", http.Dir(filesDir))

		r.Get("/", http.RedirectHandler("/dashboard", http.StatusMovedPermanently).ServeHTTP)
	})

	http.ListenAndServe(":8080", r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

// status accepts current state from minions and responds with expected state
func status(w http.ResponseWriter, r *http.Request) {
	// id := chi.URLParam(r, "id")
	// r.ParseForm()
	// query := r.Form.Get("query")
	// http.Error(w, fmt.Sprintf("failed to get client %v", err), 400)
	// w.Write(eventListJSON)
}

// kubeconfig downloads kubeconfig
func kubeconfig(w http.ResponseWriter, r *http.Request) {
	// id := chi.URLParam(r, "id")

	// http.Error(w, fmt.Sprintf("failed to get client %v", err), 400)
	// w.Write(eventListJSON)
}
