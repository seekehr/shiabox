package internal

import "server/internal/vector"

// Handler Basically handles embedding and everything.
type Handler struct {
	VectorDb *vector.Db
}

func NewHandler(vectorDb *vector.Db) *Handler {
	return &Handler{
		VectorDb: vectorDb,
	}
}
