package server

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/nodeman"
	"gitlab.com/dataptive/styx/server/config"
	"gitlab.com/dataptive/styx/server/logs_routes"
	"gitlab.com/dataptive/styx/server/nodes_routes"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

type Router struct {
	router http.Handler
	config config.Config
}

func NewRouter(logManager *logman.LogManager, nodeManager *nodeman.NodeManager, config config.Config) (r *Router) {

	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	r = &Router{
		router: router,
		config: config,
	}

	logs_routes.RegisterRoutes(router.PathPrefix("/logs").Subrouter(), logManager, config)
	nodes_routes.RegisterRoutes(router.PathPrefix("/nodes").Subrouter(), nodeManager, config)

	router.Handle("/metrics", promhttp.Handler())

	c := cors.New(cors.Options{
		AllowedOrigins:   r.config.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           0,
	})

	router.Use(c.Handler)

	return r
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	r.router.ServeHTTP(rw, req)
}

// TODO: Panic handler?

func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	api.WriteError(w, http.StatusNotFound, api.ErrNotFound)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {

	api.WriteError(w, http.StatusMethodNotAllowed, api.ErrMethodNotAllowed)
}
