package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var rawOut bytes.Buffer
	cmd.Stdout = &rawOut
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type ProbeResult struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	var probe ProbeResult
	if err = json.Unmarshal(rawOut.Bytes(), &probe); err != nil {
		return "", err
	}
	if len(probe.Streams) == 0 {
		return "", fmt.Errorf("no video streams found")
	}

	w := probe.Streams[0].Width
	h := probe.Streams[0].Height
	ratio := float64(w) / float64(h)

	const tolerance = 0.01

	if math.Abs(ratio - 16.0/9.0) < tolerance {
		return "16:9", nil
	}
	if math.Abs(ratio - 9.0/16.0) < tolerance {
		return "9:16", nil
	}
	return "other", nil
}
