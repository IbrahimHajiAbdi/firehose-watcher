package core

import (
	"context"
	"encoding/json"
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"log/slog"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
)

type PostDetails struct {
	Handle   string
	Text     string
	Repo     string
	Response *bsky.FeedPost
	Rkey     string
	Media    *utils.Media
}

type DownloadClient interface {
	FetchPostIdentifier(ctx context.Context, client api.APIClient, repo, path string) (string, error)
	FetchPostDetails(ctx context.Context, client api.APIClient, atUri string) (*PostDetails, error)
	DownloadBlobs(ctx context.Context, APIClient api.APIClient, FSClient utils.FileSystem, media *utils.Media, postDetails *PostDetails, directory string) error
}

type DefaultDownloadClient struct{}

func (dc *DefaultDownloadClient) FetchPostIdentifier(ctx context.Context, client api.APIClient, repo, path string) (string, error) {
	return FetchPostIdentifier(ctx, client, repo, path)
}

func (dc *DefaultDownloadClient) FetchPostDetails(ctx context.Context, client api.APIClient, atUri string) (*PostDetails, error) {
	return FetchPostDetails(ctx, client, atUri)
}

func (dc *DefaultDownloadClient) DownloadBlobs(ctx context.Context, APIClient api.APIClient, FSClient utils.FileSystem, media *utils.Media, postDetails *PostDetails, directory string) error {
	return DownloadBlobs(ctx, APIClient, FSClient, media, postDetails, directory)
}

func DownloadPost(ctx context.Context, downloadClient DownloadClient, APIClient api.APIClient, FSClient utils.FileSystem, repo string, repo_path string, directory string) {
	atUri, err := downloadClient.FetchPostIdentifier(ctx, APIClient, repo, repo_path)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("retrieved post aturi", "aturi", atUri)

	postDetails, err := downloadClient.FetchPostDetails(ctx, APIClient, atUri)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("retrieved post details", "details", postDetails)

	if postDetails.Media != nil {
		media := postDetails.Media

		err = downloadClient.DownloadBlobs(ctx, APIClient, FSClient, media, postDetails, directory)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("downloaded blobs associated with post", "aturi", atUri)
	}

	filename := utils.MakeFilepath(directory, postDetails.Rkey, postDetails.Handle, postDetails.Text, "json", 0, 255)

	bytes, err := json.MarshalIndent(postDetails.Response, "", "	")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	err = utils.WriteFile(FSClient, filename, &bytes)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("wrote to file system post metadata and blob(s) associated with post", "aturi", atUri)
}

func DownloadBlobs(ctx context.Context, APIClient api.APIClient, FSClient utils.FileSystem, media *utils.Media, postDetails *PostDetails, directory string) error {
	if media.ImageCid != nil {
		for i, imageCid := range media.ImageCid {
			res, err := api.GetBlob(APIClient, postDetails.Repo, imageCid)
			if err != nil {
				return err
			}
			number := 0
			if len(media.ImageCid) > 1 {
				number = i + 1
			}
			filename := utils.MakeFilepath(
				directory,
				postDetails.Rkey,
				postDetails.Handle,
				postDetails.Text,
				postDetails.Media.MediaType,
				number,
				255,
			)
			if err := utils.WriteFile(FSClient, filename, res); err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	if media.VideoCid != "" {
		res, err := api.GetBlob(APIClient, postDetails.Repo, media.VideoCid)
		if err != nil {
			return err
		}
		filename := utils.MakeFilepath(
			directory,
			postDetails.Rkey,
			postDetails.Handle,
			postDetails.Text,
			postDetails.Media.MediaType,
			0,
			255,
		)
		if err := utils.WriteFile(FSClient, filename, res); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
