package graph

import (
	"github.com/22Fariz22/forum/internal/repository"
	"github.com/22Fariz22/forum/pubsub"
)

// Repository определён в пакете repository
type Repository = repository.Repository

// Resolver содержит ссылки на хранилище и систему pubsub для подписок.
type Resolver struct {
	Repo   Repository
	PubSub *pubsub.PubSub
}

func NewResolver(repo Repository) *Resolver {
	return &Resolver{
		Repo:   repo,
		PubSub: pubsub.NewPubSub(),
	}
}
