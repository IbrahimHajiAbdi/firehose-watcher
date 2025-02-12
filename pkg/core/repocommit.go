package core

import (
	"context"
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"log/slog"
	"strings"
	"sync"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
)

func RepoCommit(
	did *atproto.IdentityResolveHandle_Output,
	directory string,
	APIClient api.APIClient,
	FSClient utils.FileSystem,
	downloadClient DownloadClient,
	semaphore *chan struct{},
	wg *sync.WaitGroup,
) *events.RepoStreamCallbacks {
	return &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo != did.Did {
				return nil
			}
			for _, op := range evt.Ops {
				if op.Action == "create" && strings.Contains(op.Path, "feed") {
					wg.Add(1)
					go func(path string) {
						(*semaphore) <- struct{}{}
						defer func() { <-(*semaphore) }()
						defer wg.Done()
						DownloadPost(context.Background(), downloadClient, APIClient, FSClient, evt.Repo, path, directory)
					}(op.Path)
				} else {
					slog.Info("Operation received", "action", op.Action, "path", op.Path)
				}
			}
			return nil
		},
	}
}
