package server

import (
	"net/http"

	"github.com/22Fariz22/forum/config"
	"github.com/22Fariz22/forum/graph"
	"github.com/22Fariz22/forum/pkg/logger"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

// Server struct
type Server struct {
	logger   logger.Logger
	cfg      *config.Config
	resolver *graph.Resolver
}

// NewServer New Server constructor
func NewServer(logger logger.Logger, cfg *config.Config, resolver *graph.Resolver) *Server {
	return &Server{logger: logger, cfg: cfg, resolver: resolver}
}

func (s *Server) Run() {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: s.resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	addr := ":" + s.cfg.Server.Port

	s.logger.Infof("Сервер запущен на %s", addr)
	s.logger.Error(http.ListenAndServe(addr, nil))
}
