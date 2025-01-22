package api

import (
	"context"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/cenkalti/backoff/v5"
)

type APIClient interface {
	SyncGetBlob(ctx context.Context, client *xrpc.Client, cid, repo string) ([]byte, error)
	RepoGetRecord(ctx context.Context, client *xrpc.Client, cid, collection, repo, rkey string) (*atproto.RepoGetRecord_Output, error)
	FeedGetPosts(ctx context.Context, client *xrpc.Client, uris []string) (*bsky.FeedGetPosts_Output, error)
}

type DefaultAPIClient struct{}

func (d *DefaultAPIClient) SyncGetBlob(ctx context.Context, client *xrpc.Client, cid, repo string) ([]byte, error) {
	return atproto.SyncGetBlob(ctx, client, cid, repo)
}

func (d *DefaultAPIClient) RepoGetRecord(ctx context.Context, client *xrpc.Client, cid, collection, repo, rkey string) (*atproto.RepoGetRecord_Output, error) {
	return atproto.RepoGetRecord(ctx, client, cid, collection, repo, rkey)
}

func (d *DefaultAPIClient) FeedGetPosts(ctx context.Context, client *xrpc.Client, uris []string) (*bsky.FeedGetPosts_Output, error) {
	return bsky.FeedGetPosts(ctx, client, uris)
}

var (
	BackoffOpts = backoff.WithBackOff(
		&backoff.ExponentialBackOff{
			InitialInterval:     1 * time.Second,
			RandomizationFactor: 0.5,
			Multiplier:          2,
			MaxInterval:         32 * time.Second,
		})
	MaxRetries = backoff.WithMaxTries(5)
)

func GetBlob(client APIClient, repo, cid string) (*[]byte, error) {
	operation := func() (*[]byte, error) {
		ctx := context.Background()
		res, err := client.SyncGetBlob(ctx, &xrpc.Client{
			Host: "https://bsky.social",
		}, cid, repo)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
	res, err := backoff.Retry(context.TODO(), operation, BackoffOpts, MaxRetries)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetRecord(client APIClient, collection, repo, rkey string) (*atproto.RepoGetRecord_Output, error) {
	operation := func() (*atproto.RepoGetRecord_Output, error) {
		ctx := context.Background()
		res, err := client.RepoGetRecord(ctx, &xrpc.Client{
			Host: "https://bsky.social",
		}, "", collection, repo, rkey)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	res, err := backoff.Retry(context.TODO(), operation, BackoffOpts, MaxRetries)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetPost(client APIClient, atUri string) (*bsky.FeedGetPosts_Output, error) {
	operation := func() (*bsky.FeedGetPosts_Output, error) {
		ctx := context.Background()
		res, err := client.FeedGetPosts(ctx, &xrpc.Client{
			Host: "https://public.api.bsky.app",
		}, []string{atUri})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	res, err := backoff.Retry(context.TODO(), operation, BackoffOpts, MaxRetries)
	if err != nil {
		return nil, err
	}
	return res, nil
}
