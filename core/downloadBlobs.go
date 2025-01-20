package core

import (
	"encoding/json"
	"firehose/api"
	"firehose/utils"
	"fmt"
	"time"
)

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
