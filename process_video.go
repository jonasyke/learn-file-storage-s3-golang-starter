package main

import "os/exec"

func processVideoForFastStart(filePath string) (string, error) {
	newFilePath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", newFilePath)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return newFilePath, nil
}
