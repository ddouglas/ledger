package internal

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/lestrrat-go/jwx/jwt"
)

type contextKey string

var CtxToken contextKey = contextKey("token")
var CtxUser contextKey = contextKey("user")

func UserFromContext(ctx context.Context) *ledger.User {
	return ctx.Value(CtxUser).(*ledger.User)
}

func TokenFromContext(ctx context.Context) jwt.Token {
	return ctx.Value(CtxToken).(jwt.Token)
}
