package cmd

import (
	"context"
	"firehose/pkg/core"
	"firehose/pkg/utils"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const (
	MAX_WORKERS = 4
)

var (
	handle string
)

// TODO: logging
var rootCmd = &cobra.Command{
	Use:   "fw --handle <handle> <directory>",
	Short: "fw is a way to subscribe to a repo and download all likes, reposts and posts on Bluesky social media as it is committed to the repo.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]
		if _, err := os.Stat(directory); err != nil {
			fmt.Println(err)
			return
		}

		f, err := utils.MakeLogFile(directory)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		utils.SetupLogger(f)

		uri := "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"
		con, _, err := websocket.DefaultDialer.Dial(uri, http.Header{})
		if err != nil {
			fmt.Println(err)
			return
		}

		client := utils.DefaultHandleResolver{}
		did, err := utils.ResolveHandle(&client, handle)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Now subscribed to:", handle)

		semaphore := make(chan struct{}, MAX_WORKERS)
		var wg sync.WaitGroup

		rsc := core.RepoCommit(did, &directory, &semaphore, &wg)

		sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)
		events.HandleRepoStream(context.Background(), con, sched, nil)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&handle, "handle", "", "Handle of the desired account")
	rootCmd.MarkPersistentFlagRequired("handle")
}
