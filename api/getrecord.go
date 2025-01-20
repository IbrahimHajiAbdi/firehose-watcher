package api

import (
	"context"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
)

func GetRecord(collection string, repo string, rkey string) (*atproto.RepoGetRecord_Output, error) {
	ctx := context.Background()
	res, err := atproto.RepoGetRecord(ctx, &xrpc.Client{
		Host: "https://bsky.social",
	}, "", collection, repo, rkey)
	if err != nil {
		return nil, err
	}
	return res, nil
}
