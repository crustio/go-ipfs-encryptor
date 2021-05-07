package crust

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
)

const sealContextKey string = "sckey"

func GenSealContext(ctx context.Context, root cid.Cid) context.Context {
	return context.WithValue(ctx, sealContextKey, root.String())
}

func GetRootFromSealContext(ctx context.Context) (cid.Cid, error) {
	v := ctx.Value(sealContextKey)
	if buf, ok := v.([]byte); ok {
		return cid.Parse(buf)
	}
	return cid.Undef, fmt.Errorf("Can't find root cid from context")
}
