package paraU

import (
	"context"
)

type paraUCtxKeyType string

const paraUCtxKey paraUCtxKeyType = "paraU"

func SetInContext(ctx context.Context, c ParaU) context.Context {
	return context.WithValue(ctx, paraUCtxKey, c)
}

func GetFromContext(ctx context.Context) ParaU {
	value := ctx.Value(paraUCtxKey)
	return value.(ParaU)
}
