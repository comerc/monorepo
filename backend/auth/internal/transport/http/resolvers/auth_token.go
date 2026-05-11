package resolvers

import "context"

type authTokenKey struct{}

// WithAuthToken кладёт bearer token в контекст GraphQL-запроса.
func (r *Resolver) WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey{}, token)
}

func authTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(authTokenKey{}).(string)
	return token
}
