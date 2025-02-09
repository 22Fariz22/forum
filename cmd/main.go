package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/22Fariz22/forum/graph"
	"github.com/22Fariz22/forum/repository"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	storageType := flag.String("storage", "inmemory", "Тип хранилища: memory или postgres")
	addr := flag.String("addr", ":8080", "Адрес сервера")
	flag.Parse()

	var repo graph.Repository

	if *storageType == "postgres" {
		// Для PostgreSQL ожидается, что DATABASE_URL задан в переменной окружения
		p, err := repository.NewPostgresRepository(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("Не удалось подключиться к PostgreSQL: %v", err)
		}
		repo = p
	} else {
		repo = repository.NewInMemoryRepository()
	}

	// Инициализируем резолвер с хранилищем и системой pubsub для подписок
	resolver := graph.NewResolver(repo)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Сервер запущен на %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
