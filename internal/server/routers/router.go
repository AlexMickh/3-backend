package routers

import "github.com/go-chi/chi/v5"

type Router interface {
	RegisterRoute(r *chi.Mux)
}
