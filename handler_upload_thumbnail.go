package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20

	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
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
	if fileType != "image/jpeg" && fileType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Only JPEG or PNG images are allowed", nil)
		return
	}

	extension := strings.Split(fileType, "/")

	key := make([]byte, 32)
	rand.Read(key)

	var rawURLEncoding = base64.RawURLEncoding
	encodedName := rawURLEncoding.EncodeToString(key)

	newFilePath := filepath.Join(cfg.assetsRoot, encodedName+"."+extension[1])

	newFile, err := os.Create(newFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create thumbnail Path", err)
		return
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy file contents", err)
		return
	}

	newURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, encodedName, extension[1])

	userVideo.ThumbnailURL = &newURL

	err = cfg.db.UpdateVideo(userVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not Update Video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, userVideo)
}
