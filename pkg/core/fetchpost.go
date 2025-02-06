package core

import (
	"encoding/json"
	"firehose/pkg/api"
	"firehose/pkg/utils"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
)

func FetchPostIdentifier(client api.APIClient, repo, path string) (string, error) {
	rkey := utils.FindExpression("[^/]*$", path)
	collection := utils.FindExpression("^[^/]*", path)

	res, err := api.GetRecord(client, collection, repo, rkey)
	if err != nil {
		return "", err
	}

	bytes, err := res.Value.MarshalJSON()
	if err != nil {
		return "", err
	}

	var postDetails bsky.FeedLike

	err = json.Unmarshal(bytes, &postDetails)
	if err != nil {
		return "", err
	}

	return postDetails.Subject.Uri, nil
}

func FetchPostDetails(client api.APIClient, atUri string) (*PostDetails, error) {
	res, err := api.GetPost(client, atUri)
	if err != nil {
		return nil, fmt.Errorf("error occured, post is either missing or deleted: %w with the ATURI: %s", err, atUri)
	}

	var postDetails PostDetails
	var record bsky.FeedPost

	if len(res.Posts) < 1 {
		return nil, fmt.Errorf("there is no post with this AT-URI: %s", atUri)
	}

	post := res.Posts[0]
	postDetails.Handle = post.Author.Handle

	rkey := utils.FindExpression("[^/]*$", post.Uri)
	postDetails.Rkey = rkey

	bytes, err := post.Record.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error occured while marshaling post to JSON: %w The post ATURI: %s", err, atUri)
	}

	err = json.Unmarshal(bytes, &record)
	if err != nil {
		return nil, fmt.Errorf("error occured while unmarshaling post JSON to type bsky.FeedPost: %w The post ATURI: %s", err, atUri)
	}

	postDetails.Text = record.Text
	postDetails.Response = &record
	postDetails.Repo = post.Author.Did

	if record.Embed != nil {
		postDetails.Media = utils.ExtractMedia(record.Embed)
	}

	return &postDetails, nil
}
