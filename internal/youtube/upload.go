package youtube

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

func UploadVideo(ctx context.Context, videoPath, title, description, thumbPath, privacy string) (string, error) {
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET_JSON")
	tokenJSON := os.Getenv("YOUTUBE_TOKEN_JSON")

	var config oauth2.Config
	if err := json.Unmarshal([]byte(clientSecret), &config); err != nil {
		return "", fmt.Errorf("parse client secret: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}

	httpClient := config.Client(ctx, &token)
	service, err := youtube.New(httpClient)
	if err != nil {
		return "", fmt.Errorf("new youtube client: %w", err)
	}

	if privacy == "" {
		privacy = "unlisted"
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "22",
		},
		Status: &youtube.VideoStatus{PrivacyStatus: privacy},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(videoPath)
	if err != nil {
		return "", fmt.Errorf("open video: %w", err)
	}
	defer file.Close()

	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("upload video: %w", err)
	}

	videoID := response.Id
	url := fmt.Sprintf("https://youtu.be/%s", videoID)

	if thumbPath != "" {
		thumbFile, err := os.Open(thumbPath)
		if err == nil {
			defer thumbFile.Close()
			service.Thumbnails.Set(videoID).Media(thumbFile).Do()
		}
	}

	return url, nil
}
