package cmd

import (
	"context"
	"firehose/pkg/api"
	"firehose/pkg/core"
	"firehose/pkg/utils"
	"fmt"
	"log/slog"
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

var rootCmd = &cobra.Command{
	Use:   "fw --handle <handle> <directory>",
	Short: "fw is a CLI tool to subscribe to a repo and download likes, reposts and posts as they are committed.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]
		if _, err := os.Stat(directory); err != nil {
			slog.Error("Directory does not exist", "error", err)
			return
		}

		f, err := utils.MakeLogFile(directory)
		if err != nil {
			slog.Error("Error creating log file", "error", err)
			return
		}
		defer f.Close()
		utils.SetupLogger(f)

		uri := "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"
		con, _, err := websocket.DefaultDialer.Dial(uri, http.Header{})
		if err != nil {
			slog.Error("WebSocket dial error", "error", err)
			return
		}

		client := utils.DefaultHandleResolver{}
		did, err := utils.ResolveHandle(&client, handle)
		if err != nil {
			slog.Error("Error resolving handle", "error", err)
			return
		}

		fmt.Println("Now subscribed to:", handle)

		semaphore := make(chan struct{}, MAX_WORKERS)
		var wg sync.WaitGroup

		APIClient := api.DefaultAPIClient{}
		FSClient := utils.DefaultFileSystem{}
		DownlaodClient := core.DefaultDownloadClient{}

		rsc := core.RepoCommit(did, directory, &APIClient, &FSClient, &DownlaodClient, &semaphore, &wg)

		sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)
		events.HandleRepoStream(context.Background(), con, sched, slog.Default())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&handle, "handle", "", "Handle of the desired account")
	rootCmd.MarkPersistentFlagRequired("handle")
}
