package core

import (
	"encoding/json"
	"firehose/api"
	"fmt"
	"regexp"

	"github.com/bluesky-social/indigo/api/bsky"
)

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
