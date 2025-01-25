package cmd

import (
	"context"
	"firehose/pkg/core"
	"firehose/pkg/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	handle    string
	directory string
)

// TODO: Add tests
// TODO: Add graceful failure and logging what failed
var rootCmd = &cobra.Command{
	Use:   "fw",
	Short: "fw is a way to subscribe to a repo and download all likes, reposts and posts on Bluesky social media as it is committed to the repo",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(directory); err != nil {
			fmt.Println(err)
			return
		}

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

		fmt.Println("Now subsribed to:", handle)

		rsc := core.RepoCommit(did, directory)

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
	rootCmd.PersistentFlags().StringVar(&directory, "directory", "", "Directory to download media")
	rootCmd.MarkPersistentFlagRequired("handle")
	rootCmd.MarkPersistentFlagRequired("directory")
}
