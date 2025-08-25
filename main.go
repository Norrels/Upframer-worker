package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	videoPath := os.Args[1]

	outputDir := "frames"

	err := os.MkdirAll(outputDir, 0755)

	if err != nil {
		log.Fatal("Error creating folder", outputDir)
	}

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		outputDir+"/frame_%04d.jpg",
	)

	_, err = cmd.CombinedOutput()

	if err != nil {
		log.Fatal("Error extracting frames", err)
	}

	fmt.Println("Frames successfully extracted in the folder:", outputDir)
}
