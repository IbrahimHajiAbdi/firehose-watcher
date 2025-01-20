package api

import (
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

func GetBlob(repo string, cid string) (*[]byte, error) {
	ctx := context.Background()
	res, err := atproto.SyncGetBlob(ctx, &xrpc.Client{
		Host: "https://bsky.social",
	}, cid, repo)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
