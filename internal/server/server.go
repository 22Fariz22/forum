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
	// serverAddr := fmt.Sprintf("%s:%s", s.cfg.Server.BaseUrl, s.cfg.Server.Port)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: s.resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	s.logger.Infof("Сервер запущен на %s", ":8080")
	s.logger.Error(http.ListenAndServe(":8080", nil))

	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	//
	// <-quit
	//
	// ctx, shutdown := context.WithTimeout(context.Background(), s.cfg.Server.CtxTimeout)
	// defer shutdown()
	//
	// s.logger.Info("Server Exited Properly")
	// return s.echo.Server.Shutdown(ctx)
}
