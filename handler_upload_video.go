package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	defer r.Body.Close()

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not retrieve video ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve video", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	userVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve video", err)
		return
	}
	if userVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	fileType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse media type", err)
		return
	}

	if fileType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "wrong video format, must be mp4", err)
		return
	}

	tempVideo, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not process video", err)
		return
	}
	defer os.Remove(tempVideo.Name())
	defer tempVideo.Close()

	if _, err := io.Copy(tempVideo, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy file", err)
		return
	}

	tempVideo.Seek(0, io.SeekStart)

	key, err := getAssetPath(fileType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate key", err)
		return
	}

	aspect, err := getVideoAspectRatio(tempVideo.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest,"could not retrieve aspect ratio", err)
		return
	}

	switch aspect {
	case "16:9":
		key = "landscape/" + key
	case "9:16":
		key = "portrait/" + key
	default:
		key = "other/" + key
	}

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(key),
		Body:        tempVideo,
		ContentType: aws.String(fileType),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to put object", err)
		return
	}

	newVideoURL := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, key)
	userVideo.VideoURL = &newVideoURL

	err = cfg.db.UpdateVideo(userVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update video", err)
		return
	}

	respondWithJSON(w, 200, userVideo)
}
