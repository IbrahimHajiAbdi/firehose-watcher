package core

import (
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"fmt"
	"strings"
	"sync"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
)

func RepoCommit(did *atproto.IdentityResolveHandle_Output, directory *string, semaphore *chan struct{}, wg *sync.WaitGroup) *events.RepoStreamCallbacks {
	APIClient := api.DefaultAPIClient{}
	FSClient := utils.DefaultFileSystem{}
	DownloadClient := DefaultDownloadClient{}
	var rsc = &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo != did.Did {
				return nil
			}

			if evt.Ops[0].Action == "create" && strings.Contains(evt.Ops[0].Path, "feed") {
				wg.Add(1)

				go func() {
					defer wg.Done()

					*semaphore <- struct{}{}
					defer func() { <-*semaphore }()

					DownloadPost(&DownloadClient, &APIClient, &FSClient, evt.Repo, evt.Ops[0].Path, *directory)
				}()
			}

			for _, op := range evt.Ops {
				fmt.Printf(" - %s record %s\n", op.Action, op.Path)
			}

			return nil
		},
	}
	return rsc
}
