package api

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

func GetPost(atUri string) (*bsky.FeedGetPosts_Output, error) {
	ctx := context.Background()
	res, err := bsky.FeedGetPosts(ctx, &xrpc.Client{
		Host: "https://public.api.bsky.app",
	}, []string{atUri})
	if err != nil {
		return nil, err
	}
	return res, nil
}
