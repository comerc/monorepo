package http

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	transportgraphql "github.com/pure-golang/monorepo/backend/profile/internal/transport/http/graphql"
	"github.com/pure-golang/monorepo/backend/profile/internal/transport/http/resolvers"
)

// Transport регистрирует HTTP-маршруты profile-сервиса.
type Transport struct {
	resolver *resolvers.Resolver
}

// New создаёт HTTP-транспорт profile-сервиса.
func New(resolver *resolvers.Resolver) *Transport {
	return &Transport{resolver: resolver}
}

// EnrichRoutes регистрирует маршруты транспорта в mux.
func (t *Transport) EnrichRoutes(mux *http.ServeMux) {
	srv := handler.NewDefaultServer(
		transportgraphql.NewExecutableSchema(transportgraphql.Config{Resolvers: t.resolver}),
	)
	mux.Handle("POST /graphql", t.withUserID(srv))
	mux.Handle("GET /playground", playground.Handler("Profile GraphQL", "/graphql"))
}

func (t *Transport) withUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(t.resolver.WithUserID(r.Context(), r.Header.Get("X-User-ID"))))
	})
}
