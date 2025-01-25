package core

import (
	"encoding/json"
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
)

func fetchPostIdentifier(client api.APIClient, repo, path string) (string, error) {
	rkey := utils.FindExpression("[^/]*$", path)
	collection := utils.FindExpression("^[^/]*", path)

	res, err := api.GetRecord(client, collection, repo, rkey)
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

func fetchPostDetails(client api.APIClient, atUri string) (*PostDetails, error) {
	res, err := api.GetPost(client, atUri)
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

	rkey := utils.FindExpression("[^/]*$", post.Uri)
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
		postDetails.Media = utils.ExtractMedia(record.Embed)
	}

	return &postDetails, nil
}
