package utils

import (
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

func ResolveHandle(handle string) (*atproto.IdentityResolveHandle_Output, error) {
	did, err := atproto.IdentityResolveHandle(context.Background(), &xrpc.Client{
		Host: "https://bsky.social",
	}, handle)
	if err != nil {
		return nil, err
	}
	return did, nil
}
