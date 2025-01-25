package utils

import (
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

type HandleResolver interface {
	IdentityResolveHandle(ctx context.Context, c *xrpc.Client, handle string) (*atproto.IdentityResolveHandle_Output, error)
}

type DefaultHandleResolver struct{}

func (d *DefaultHandleResolver) IdentityResolveHandle(ctx context.Context, c *xrpc.Client, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	return atproto.IdentityResolveHandle(ctx, c, handle)
}

func ResolveHandle(client HandleResolver, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	did, err := client.IdentityResolveHandle(context.Background(), &xrpc.Client{
		Host: "https://bsky.social",
	}, handle)
	if err != nil {
		return nil, err
	}
	return did, nil
}
