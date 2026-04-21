package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	newSignedClient := s3.NewPresignClient(s3Client)
	eSClient, err := newSignedClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	},
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", err
	}

	return eSClient.URL, nil

}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	var videoURL []string
	if video.VideoURL == nil {
		return video, nil
	}
	videoURL = strings.Split(*video.VideoURL, ",")
	if len(videoURL) != 2 {
		return database.Video{}, fmt.Errorf("invalid video url format")
	}

	timeValid := 15 * time.Minute
	presignedURL, err := generatePresignedURL(cfg.s3Client, videoURL[0], videoURL[1], timeValid)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &presignedURL
	return video, nil
}
