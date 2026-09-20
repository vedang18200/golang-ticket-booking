package graph

import (
	"context"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"ticket-booking/graph/model"
	"ticket-booking/internal/auth"
)

// codedErr returns a GraphQL error with extensions.code set (clients and k6 branch on it).
func codedErr(code, msg string) error {
	return &gqlerror.Error{Message: msg, Extensions: map[string]any{"code": code}}
}

// internalErr logs the real error and hides it from the client.
func internalErr(err error) error {
	slog.Error("internal error", "err", err)
	return codedErr("INTERNAL", "internal error")
}

func authPayload(u *auth.User, token string) *model.AuthPayload {
	return &model.AuthPayload{
		Token: token,
		User:  &model.User{ID: u.ID, Email: u.Email, Role: model.Role(u.Role)},
	}
}

// HasRole implements the @hasRole directive. ADMIN can do everything USER can.
func HasRole(ctx context.Context, _ any, next graphql.Resolver, role model.Role) (any, error) {
	id, ok := auth.FromContext(ctx)
	if !ok {
		return nil, codedErr("UNAUTHENTICATED", "login required")
	}
	if role == model.RoleAdmin && id.Role != string(model.RoleAdmin) {
		return nil, codedErr("FORBIDDEN", "admin only")
	}
	return next(ctx)
}
