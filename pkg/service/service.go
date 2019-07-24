package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	v1 "github.com/gtracer/overlord/api/v1"
	"github.com/gtracer/overlord/pkg/cluster"
	"github.com/gtracer/overlord/pkg/minion"
	"k8s.io/client-go/kubernetes/scheme"
)

// Run ...
func Run() {
	v1.AddToScheme(scheme.Scheme)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "dashboard")
	FileServer(r, "/dashboard", http.Dir(filesDir))

	r.Get("/", http.RedirectHandler("/dashboard", http.StatusMovedPermanently).ServeHTTP)
	r.Get("/clusters", listClusters)

	r.Mount("/{userid}", CustomerRoutes())

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

// CustomerRoutes creates a REST router for the for a customer
func CustomerRoutes() chi.Router {
	r := chi.NewRouter()
	// r.Use() // some middleware..

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/minions", listMinions)
		r.Get("/kubeconfig", kubeconfig)
		r.Post("/bootstrap", report)
		r.Post("/{minionid}", report)
	})

	return r
}

func listClusters(w http.ResponseWriter, r *http.Request) {
	jsonData, err := cluster.List()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list clusters %v", err), 400)
		return
	}
	w.Write(jsonData)
}

// report accepts current state from minions and responds with expected state
func report(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	id := chi.URLParam(r, "id")
	minionID := chi.URLParam(r, "minionid")

	if err := cluster.Report(userID, id); err != nil {
		http.Error(w, fmt.Sprintf("failed to report cluster %v", err), 400)
		return
	}

	minionStatus, err := minion.Status(id, minionID, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get minion status %v", err), 400)
		return
	}

	jsonData, err := minion.Report(userID, id, minionID, minionStatus)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get config %v", err), 400)
		return
	}

	w.Write(jsonData)
}

func listMinions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	id := chi.URLParam(r, "id")

	jsonData, err := minion.List(userID, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list minions %v", err), 400)
		return
	}
	w.Write(jsonData)
}

// kubeconfig downloads kubeconfig
func kubeconfig(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	id := chi.URLParam(r, "id")

	kubeconfig, err := cluster.Kubeconfig(userID, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get kubeconfig %v", err), 400)
		return
	}
	w.Write(kubeconfig)
}

// bootstrap enables bootstrap on cluster
func bootstrap(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	id := chi.URLParam(r, "id")

	if err := cluster.Bootstrap(userID, id); err != nil {
		http.Error(w, fmt.Sprintf("failed to report cluster %v", err), 400)
		return
	}
}
