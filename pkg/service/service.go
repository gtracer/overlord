package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	v1 "github.com/gtracer/overlord/api/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Run ...
func Run() {
	v1.AddToScheme(scheme.Scheme)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

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

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "dashboard")
	FileServer(r, "/dashboard", http.Dir(filesDir))

	r.Get("/", http.RedirectHandler("/dashboard", http.StatusMovedPermanently).ServeHTTP)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/kubeconfig", kubeconfig)
	})

	r.Route("/{nodeid}", func(r chi.Router) {
		r.Post("/", report)
		r.Get("/", status)
	})

	return r
}

// report accepts current state from minions and responds with expected state
func report(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	id := chi.URLParam(r, "id")
	minionID := chi.URLParam(r, "minionid")
	cfg, err := ctrlcfg.GetConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get config %v", err), 400)
		return
	}

	client, err := ctrlclient.New(cfg, ctrlclient.Options{})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get client %v", err), 400)
		return
	}

	clusterName := fmt.Sprintf("%s_%s_%s", userID, id, minionID)
	cluster := &v1.Cluster{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterName}, cluster)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get cluster %s, %v", clusterName, err), 400)
		return
	}

	// r.ParseForm()
	// query := r.Form.Get("query")
	// http.Error(w, fmt.Sprintf("failed to get client %v", err), 400)
	// w.Write(eventListJSON)
}

func status(w http.ResponseWriter, r *http.Request) {
}

// kubeconfig downloads kubeconfig
func kubeconfig(w http.ResponseWriter, r *http.Request) {
	// id := chi.URLParam(r, "id")

	// http.Error(w, fmt.Sprintf("failed to get client %v", err), 400)
	// w.Write(eventListJSON)
}
