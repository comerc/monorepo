package resolvers

import "context"

type userIDKey struct{}

// WithUserID кладёт user_id в контекст GraphQL-запроса.
func (r *Resolver) WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func userIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}
