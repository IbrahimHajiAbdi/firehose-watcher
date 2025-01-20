package utils

import (
	"regexp"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Media struct {
	Video_Cid string
	Image_Cid []string
	MediaType string
}

func ExtractMedia(record *bsky.FeedPost_Embed) *Media {
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
