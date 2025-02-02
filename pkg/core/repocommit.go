package core

import (
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
)

func RepoCommit(did *atproto.IdentityResolveHandle_Output, directory string) *events.RepoStreamCallbacks {
	APIClient := api.DefaultAPIClient{}
	FSClient := utils.DefaultFileSystem{}
	DownloadClient := DefaultDownloadClient{}
	var rsc = &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo != did.Did {
				return nil
			}

			if evt.Ops[0].Action == "create" && strings.Contains(evt.Ops[0].Path, "feed") {
				go DownloadPost(&DownloadClient, &APIClient, &FSClient, evt.Repo, evt.Ops[0].Path, directory)
			}

			for _, op := range evt.Ops {
				fmt.Printf(" - %s record %s\n", op.Action, op.Path)
			}

			return nil
		},
	}
	return rsc
}
