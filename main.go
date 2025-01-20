package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"firehose/api"
	"firehose/utils"
)

type PostDetails struct {
	Handle   string
	Text     string
	Repo     string
	Response *bsky.FeedPost
	Rkey     string
	Media    *Media
}

type Media struct {
	Video_Cid string
	Image_Cid []string
	MediaType string
}

func main() {
	uri := "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"
	con, _, err := websocket.DefaultDialer.Dial(uri, http.Header{})
	directory := "./media"

	err = godotenv.Load()

	if err != nil {
		fmt.Println(err)
		return
	}

	account := os.Getenv("ACCOUNT")

	did, err := atproto.IdentityResolveHandle(context.Background(), &xrpc.Client{
		Host: "https://bsky.social",
	}, account)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(did.Did)

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo != did.Did {
				return nil
			}

			if evt.Ops[0].Action == "create" {
				go downloadPost(evt.Repo, evt.Ops[0].Path, directory)
			}

			for _, op := range evt.Ops {
				fmt.Printf(" - %s record %s\n", op.Action, op.Path)
			}

			return nil
		},
	}

	sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)
	events.HandleRepoStream(context.Background(), con, sched, nil)
}

func fetchPostIdentifier(repo string, path string) (string, error) {
	re := regexp.MustCompile(`[^/]*$`)
	rkey := re.FindString(path)

	re = regexp.MustCompile(`^[^/]*`)
	collection := re.FindString(path)

	res, err := api.GetRecord(collection, repo, rkey)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	bytes, err := res.Value.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var postDetails bsky.FeedLike

	err = json.Unmarshal(bytes, &postDetails)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fetchPostDetails(postDetails.Subject.Uri)

	return postDetails.Subject.Uri, nil
}

func fetchPostDetails(atUri string) (*PostDetails, error) {
	res, err := api.GetPost(atUri)
	if err != nil {
		fmt.Println(err)
		fmt.Println(atUri)
		return nil, err
	}

	var postDetails PostDetails
	var record bsky.FeedPost

	if len(res.Posts) < 1 {
		fmt.Println("There is no post with this AT-URI: ", atUri)
		return nil, nil
	}

	post := res.Posts[0]
	postDetails.Handle = post.Author.Handle

	re := regexp.MustCompile("[^/]*$")
	rkey := re.FindString(post.Uri)
	postDetails.Rkey = rkey

	bytes, err := post.Record.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		fmt.Println(atUri)
		return nil, err
	}

	err = json.Unmarshal(bytes, &record)
	if err != nil {
		fmt.Println(err)
		fmt.Println(atUri)
		return nil, err
	}

	postDetails.Text = record.Text
	postDetails.Response = &record
	postDetails.Repo = post.Author.Did

	if record.Embed != nil {
		postDetails.Media = extractMedia(record.Embed)

	}

	return &postDetails, nil
}

// TODO: Check if the property was set, instead of empty
func extractMedia(record *bsky.FeedPost_Embed) *Media {
	extractedMedia := Media{}

	if record.EmbedRecordWithMedia != nil {
		media := record.EmbedRecordWithMedia.Media

		if media.EmbedImages != nil {
			for _, image := range media.EmbedImages.Images {
				extractedMedia.Image_Cid = append(extractedMedia.Image_Cid, image.Image.Ref.String())
			}
			mimeType := media.EmbedImages.Images[0].Image.MimeType

			re := regexp.MustCompile("[^/]*$")
			mediaType := re.FindString(mimeType)
			extractedMedia.MediaType = mediaType
		}
		if media.EmbedVideo != nil {
			extractedMedia.Video_Cid = media.EmbedVideo.Video.Ref.String()

			mimeType := record.EmbedVideo.Video.MimeType

			re := regexp.MustCompile("[^/]*$")
			mediaType := re.FindString(mimeType)
			extractedMedia.MediaType = mediaType
		}
	}
	if record.EmbedImages != nil {
		for _, image := range record.EmbedImages.Images {
			extractedMedia.Image_Cid = append(extractedMedia.Image_Cid, image.Image.Ref.String())
		}
		mimeType := record.EmbedImages.Images[0].Image.MimeType

		re := regexp.MustCompile("[^/]*$")
		mediaType := re.FindString(mimeType)
		extractedMedia.MediaType = mediaType
	}
	if record.EmbedVideo != nil {
		extractedMedia.Video_Cid = record.EmbedVideo.Video.Ref.String()

		mimeType := record.EmbedVideo.Video.MimeType

		re := regexp.MustCompile("[^/]*$")
		mediaType := re.FindString(mimeType)
		extractedMedia.MediaType = mediaType
	}

	return &extractedMedia
}

func downloadPost(repo string, repo_path string, directory string) error {
	atUri, err := fetchPostIdentifier(repo, repo_path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	postDetails, err := fetchPostDetails(atUri)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if postDetails.Media != nil {
		media := postDetails.Media

		err = downloadBlobs(media, postDetails, directory)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	filename := utils.MakeFilepath(directory, postDetails.Rkey, postDetails.Handle, postDetails.Text, "json", 255)

	bytes, err := json.Marshal(postDetails.Response)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = utils.WriteFile(filename, &bytes)
	if err != nil {
		return err
	}
	return nil
}

func downloadBlobs(media *Media, postDetails *PostDetails, directory string) error {
	filename := utils.MakeFilepath(directory, postDetails.Rkey, postDetails.Handle, postDetails.Text, postDetails.Media.MediaType, 255)
	if media.Image_Cid != nil {
		for _, imageCid := range media.Image_Cid {
			res, err := api.GetBlob(postDetails.Repo, imageCid)
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = utils.WriteFile(filename, res)
			if err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	if media.Video_Cid != "" {
		res, err := api.GetBlob(postDetails.Repo, media.Video_Cid)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = utils.WriteFile(filename, res)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
