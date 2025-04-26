package upload

import (
	uploader "download-upload/internal/upload"

	"github.com/spf13/cobra"
)

var (
	RootFolderID string
	Gdrive       uploader.Uploader
)

var UploadCmd = &cobra.Command{
	Use:  "upload-gdrive [service-account-key] [root-folder-id]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		Gdrive, err = uploader.NewGDrive(
			cmd.Context(),
			args[0],
			// "uploadvideos-455214-7ddb60eda5c2.json",
			// 15UGP04Ph8oIP5Y9TrsjjeC3a7xrYt5iq
		)

		RootFolderID = args[1]

		return
	},
}

func init() {
}
