package uploader

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type Uploader interface {
	UploadMP4(name, parentID string, f io.Reader) (string, error)
	CreateFolder(name string, parentsIDs ...string) (string, error)
}

type gDrive struct {
	srv *drive.Service
}

type UploadOptions string

var (
	Local  UploadOptions = "Local"
	GDrive UploadOptions = "Google Drive"
)

func NewGDrive(
	ctx context.Context,
	serviceAccountKey string,
) (gdrive *gDrive, err error) {
	gdrive = &gDrive{}

	gdrive.srv, err = drive.NewService(ctx, option.WithCredentialsFile(serviceAccountKey))

	return
}

func (up gDrive) CreateFolder(name string, parentsIDs ...string) (string, error) {
	folder, err := up.srv.Files.Create(&drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  parentsIDs,
	}).Do()
	if err != nil {
		return "", err
	}

	return folder.Id, nil
}

func (up gDrive) UploadMP4(name, parentID string, f io.Reader) (string, error) {
	fileMetadata := &drive.File{
		Name:    name,
		Parents: []string{parentID},
	}

	upload := up.srv.Files.Create(fileMetadata).Media(f, googleapi.ContentType("video/mp4"))
	uploadRes, err := upload.Do()
	if err != nil {
		return "", err
	}

	return uploadRes.Id, nil
}
