package core

import (
	"encoding/json"
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"fmt"
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

func DownloadPost(client *api.DefaultAPIClient, repo string, repo_path string, directory string) {
	atUri, err := fetchPostIdentifier(client, repo, repo_path)
	if err != nil {
		fmt.Println(err)
		return
	}

	postDetails, err := fetchPostDetails(client, atUri)
	if err != nil {
		fmt.Println(err)
		return
	}

	if postDetails.Media != nil {
		media := postDetails.Media

		err = downloadBlobs(client, media, postDetails, directory)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	filename := utils.MakeFilepath(directory, postDetails.Rkey, postDetails.Handle, postDetails.Text, "json", 0, 255)

	bytes, err := json.MarshalIndent(postDetails.Response, "", "	")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = utils.WriteFile(filename, &bytes)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func downloadBlobs(client *api.DefaultAPIClient, media *utils.Media, postDetails *PostDetails, directory string) error {
	if media.ImageCid != nil {
		for i, imageCid := range media.ImageCid {
			res, err := api.GetBlob(client, postDetails.Repo, imageCid)
			if err != nil {
				fmt.Println(err)
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
			err = utils.WriteFile(filename, res)
			if err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	if media.VideoCid != "" {
		res, err := api.GetBlob(client, postDetails.Repo, media.VideoCid)
		if err != nil {
			fmt.Println(err)
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
		err = utils.WriteFile(filename, res)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
