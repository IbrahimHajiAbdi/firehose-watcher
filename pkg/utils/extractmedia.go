package utils

import (
	"github.com/bluesky-social/indigo/api/bsky"
)

type Media struct {
	VideoCid  string
	ImageCid  []string
	MediaType string
}

func ExtractMedia(record *bsky.FeedPost_Embed) *Media {
	extractedMedia := Media{}

	if record.EmbedRecordWithMedia != nil {
		media := record.EmbedRecordWithMedia.Media

		if media.EmbedImages != nil {
			for _, image := range media.EmbedImages.Images {
				extractedMedia.ImageCid = append(extractedMedia.ImageCid, image.Image.Ref.String())
			}
			mimeType := media.EmbedImages.Images[0].Image.MimeType
			extractedMedia.MediaType = FindExpression("[^/]*$", mimeType)
		}
		if media.EmbedVideo != nil {
			extractedMedia.VideoCid = media.EmbedVideo.Video.Ref.String()

			mimeType := record.EmbedVideo.Video.MimeType
			extractedMedia.MediaType = FindExpression("[^/]*$", mimeType)
		}
	}
	if record.EmbedImages != nil {
		for _, image := range record.EmbedImages.Images {
			extractedMedia.ImageCid = append(extractedMedia.ImageCid, image.Image.Ref.String())
		}
		mimeType := record.EmbedImages.Images[0].Image.MimeType
		extractedMedia.MediaType = FindExpression("[^/]*$", mimeType)
	}
	if record.EmbedVideo != nil {
		extractedMedia.VideoCid = record.EmbedVideo.Video.Ref.String()

		mimeType := record.EmbedVideo.Video.MimeType
		extractedMedia.MediaType = FindExpression("[^/]*$", mimeType)
	}

	return &extractedMedia
}
