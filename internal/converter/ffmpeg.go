package converter

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

type Converter interface {
	TsToMP4(f io.Reader) (io.Reader, error)
}

type ffmpeg struct {
}

func NewFfmpeg() ffmpeg {
	return ffmpeg{}
}

func (c ffmpeg) TsToMP4(f io.Reader) (io.Reader, error) {
	inputFile, err := os.CreateTemp("", "input-*.ts")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	if _, err = io.Copy(inputFile, f); err != nil {
		return nil, err
	}

	outputFile, err := os.CreateTemp("", "output-*.mp4")
	if err != nil {
		return nil, err
	}
	outputFileName := outputFile.Name()
	outputFile.Close()
	defer os.Remove(outputFileName)

	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile.Name(), "-c", "copy", outputFileName)

	if err = cmd.Run(); err != nil {
		return nil, err
	}

	outputData, err := os.Open(outputFileName)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, outputData)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}
