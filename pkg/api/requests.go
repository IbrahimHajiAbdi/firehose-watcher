package api

import (
	"context"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/cenkalti/backoff/v5"
)

var opts = backoff.ExponentialBackOff{
	InitialInterval:     1 * time.Second,
	RandomizationFactor: 0.5,
	Multiplier:          2,
	MaxInterval:         32 * time.Second,
}

func GetBlob(repo string, cid string) (*[]byte, error) {
	operation := func() (*[]byte, error) {
		ctx := context.Background()
		res, err := atproto.SyncGetBlob(ctx, &xrpc.Client{
			Host: "https://bsky.social",
		}, cid, repo)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
	res, err := backoff.Retry(
		context.TODO(),
		operation,
		backoff.WithBackOff(&opts),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetRecord(collection string, repo string, rkey string) (*atproto.RepoGetRecord_Output, error) {
	operation := func() (*atproto.RepoGetRecord_Output, error) {
		ctx := context.Background()
		res, err := atproto.RepoGetRecord(ctx, &xrpc.Client{
			Host: "https://bsky.social",
		}, "", collection, repo, rkey)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	res, err := backoff.Retry(
		context.TODO(),
		operation,
		backoff.WithBackOff(&opts),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetPost(atUri string) (*bsky.FeedGetPosts_Output, error) {
	operation := func() (*bsky.FeedGetPosts_Output, error) {
		ctx := context.Background()
		res, err := bsky.FeedGetPosts(ctx, &xrpc.Client{
			Host: "https://public.api.bsky.app",
		}, []string{atUri})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	res, err := backoff.Retry(
		context.TODO(),
		operation,
		backoff.WithBackOff(&opts),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}
