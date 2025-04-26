package streamings

import (
	"download-upload/cmd/upload"
	"download-upload/internal/converter"
	"download-upload/internal/download"
	uploader "download-upload/internal/upload"
	"download-upload/pkg/streamings"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	m3u8    download.Downloader
	dirPath string
)

var FullCycleCmd = &cobra.Command{
	Use: "full-cycle",
	RunE: func(cmd *cobra.Command, args []string) error {
		usernamePrompt := promptui.Prompt{
			Label:   "Enter your username",
			Pointer: promptui.PipeCursor,
		}
		username, err := usernamePrompt.Run()
		if err != nil {
			return err
		}

		passwordPrompt := promptui.Prompt{
			Label: "Enter your password",
			Mask:  '*',
		}
		password, err := passwordPrompt.Run()
		if err != nil {
			return err
		}

		fc, err := streamings.NewFullCycle(username, password)
		if err != nil {
			return err
		}

		sc := fc.GetCourseSlice()
		coursePrompt := promptui.Select{
			Label: "Select a course to download from",
			Items: sc,
			Templates: &promptui.SelectTemplates{
				Active:   "❯ {{ .Category.Name | magenta }}",
				Inactive: "   {{ .Category.Name | cyan }}",
				Selected: "{{ .Category.Name | green }}",
			},
		}
		i, _, err := coursePrompt.Run()
		if err != nil {
			return err
		}

		uploadOptionPrompt := promptui.Select{
			Label: "Select a upload option",
			Items: []uploader.UploadOptions{
				uploader.Local,
				uploader.GDrive,
			},
			Templates: &promptui.SelectTemplates{
				Active:   "❯ {{ . | magenta }}",
				Inactive: "   {{ . | cyan }}",
				Selected: "{{ . | green }}",
			},
		}
		_, selected, err := uploadOptionPrompt.Run()
		if err != nil {
			return err
		}

		switch selected {
		case string(uploader.GDrive):
			// srvAccKeyPrompt := promptui.Prompt{
			// 	Label:   "Enter the service account json file path",
			// 	Pointer: promptui.PipeCursor,
			// }
			// srvAccKey, err := srvAccKeyPrompt.Run()
			// if err != nil {
			// 	return err
			// }

			// gDriveRootFolderPrompt := promptui.Prompt{
			// 	Label:   "Enter the root dir ID",
			// 	Pointer: promptui.PipeCursor,
			// }
			// gDriveRootFolder, err := gDriveRootFolderPrompt.Run()
			// if err != nil {
			// 	return err
			// }

			err = upload.UploadCmd.RunE(
				upload.UploadCmd,
				[]string{
					"uploadvideos-455214-7ddb60eda5c2.json",
					"15UGP04Ph8oIP5Y9TrsjjeC3a7xrYt5iq",
				},
			)
			if err != nil {
				return err
			}
		case string(uploader.Local):
			dirPath = filepath.Join(".", "course")
		}

		downloadOptionPrompt := promptui.Select{
			Label: "Select a download option",
			Items: []download.DownloadOptions{
				download.AllCourse,
				download.AllModule,
				download.SpecificVideo,
			},
			Templates: &promptui.SelectTemplates{
				Active:   "❯ {{ . | magenta }}",
				Inactive: "   {{ . | cyan }}",
				Selected: "{{ . | green }}",
			},
		}
		_, selected, err = downloadOptionPrompt.Run()
		if err != nil {
			return err
		}

		switch selected {
		case string(download.AllCourse):
			err = downloadAllCourse(cmd, &fc, sc[i].Category.ID)
			if err != nil {
				return err
			}
		case string(download.AllModule):
			fmt.Println("Under construction")
		case string(download.SpecificVideo):
			err = downloadSpecificVideo(cmd, &fc, sc[i].Category.ID)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	FullCycleCmd.AddCommand(upload.UploadCmd)

	m3u8 = download.NewM3U8(
		"https://vz-13b1a7c5-e0b.b-cdn.net",
		"https://iframe.mediadelivery.net/",
		"https://iframe.mediadelivery.net/",
	)
}

func downloadAllCourse(cmd *cobra.Command, fc *streamings.FullCycle, curseID int) error {
	err := fc.GetModules(curseID)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	switch true {
	case upload.Gdrive != nil:
		for _, m := range fc.Modules {
			err := fc.GetChapters(fc.Curses[curseID].Category.ID, m.ID)
			if err != nil {
				return err
			}

			moduleFolderID, err := upload.Gdrive.CreateFolder(
				m.Name,
				upload.RootFolderID,
			)
			if err != nil {
				return err
			}

			log.Info().Str("module", m.Name).Send()
			log.Info().Int("chapters", len(fc.Chapters)).Send()

			for _, chapter := range fc.Chapters {
				chapterFolderID, err := upload.Gdrive.CreateFolder(
					chapter.Name,
					moduleFolderID,
				)
				if err != nil {
					return err
				}

				log.Info().Str("chapter", chapter.Name).Send()
				log.Info().Int("chapter-contents", len(chapter.Contents)).Send()

				for _, content := range chapter.Contents {
					wg.Add(1)

					go func(c streamings.Content, w *sync.WaitGroup) {
						defer func() {
							<-sem
							w.Done()
						}()
						sem <- struct{}{}

						log.Info().Str("video", c.Title).Send()

						f, err := m3u8.DownloadVideo(
							c.VideoID,
							c.Title,
						)
						if err != nil {
							fmt.Println(err)
							return
						}

						ffmpeg := converter.NewFfmpeg()
						log.Info().
							Str("video", c.Title).
							Str("message", "converting ts to mp4").
							Send()
						mp4, err := ffmpeg.TsToMP4(f)
						if err != nil {
							fmt.Println(err)
							return
						}
						log.Info().
							Str("video", c.Title).
							Str("message", "video converted").
							Send()

						log.Info().
							Str("video", c.Title).
							Str("message", "uploading file").
							Send()
						_, err = upload.Gdrive.UploadMP4(
							c.Title,
							chapterFolderID,
							mp4,
						)
						if err != nil {
							fmt.Println(err)
							return
						}
						log.Info().
							Str("video", c.Title).
							Str("message", "file uploaded").
							Send()
					}(content, &wg)
				}
			}

			wg.Wait()
		}
	default:
		for _, m := range fc.Modules {
			err := fc.GetChapters(fc.Curses[curseID].Category.ID, m.ID)
			if err != nil {
				return err
			}

			dirPath = filepath.Join(dirPath, m.Name)
			if err = os.MkdirAll(dirPath, 0755); err != nil {
				return err
			}

			log.Info().Str("module", m.Name).Send()
			log.Info().Int("chapters", len(fc.Chapters)).Send()

			for _, chapter := range fc.Chapters {
				filePath := filepath.Join(dirPath, chapter.Name)
				if err = os.MkdirAll(filePath, 0755); err != nil {
					return err
				}

				log.Info().Str("chapter", chapter.Name).Send()
				log.Info().Int("chapter-contents", len(chapter.Contents)).Send()

				for _, content := range chapter.Contents {
					wg.Add(1)

					go func(c streamings.Content, w *sync.WaitGroup) {
						defer func() {
							<-sem
							w.Done()
						}()
						sem <- struct{}{}

						log.Info().Str("video", content.Title).Send()

						f, err := m3u8.DownloadVideo(
							content.VideoID,
							content.Title,
						)
						if err != nil {
							fmt.Println(err)
							return
						}

						ffmpeg := converter.NewFfmpeg()
						fmt.Println("Converting TS to MP4...")
						mp4, err := ffmpeg.TsToMP4(f)
						if err != nil {
							fmt.Println(err)
							return
						}

						filePath = filepath.Join(dirPath, chapter.Name, content.Title+".mp4")
						file, err := os.Create(filePath)
						if err != nil {
							fmt.Println(err)
							return
						}
						defer file.Close()

						fmt.Println("Creating local file...")
						_, err = io.Copy(file, mp4)
						if err != nil {
							fmt.Println(err)
							return
						}
					}(content, &wg)
				}
			}

			wg.Wait()
		}
	}

	return nil
}

func downloadSpecificVideo(cmd *cobra.Command, fc *streamings.FullCycle, curseID int) error {
	err := fc.GetModules(curseID)
	if err != nil {
		return err
	}

	fmt.Println("Searching modules...")
	modulePrompt := promptui.Select{
		Label: "Select a module to download from",
		Items: fc.Modules,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Active:   "❯ {{ .Name | magenta }}",
			Inactive: "   {{ .Name | cyan }}",
			Details:  "{{ .Desc | white }}",
			Selected: "{{ .Name | green }}",
		},
	}
	moduleIdx, _, err := modulePrompt.Run()
	if err != nil {
		return err
	}

	fmt.Println("Searching chapters...")
	err = fc.GetChapters(fc.Curses[curseID].Category.ID, fc.Modules[moduleIdx].ID)
	if err != nil {
		return err
	}

	chaptersPrompt := promptui.Select{
		Label: "Select a chapter to download from",
		Items: fc.Chapters,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Active:   "❯ {{ .Name | magenta }}",
			Inactive: "   {{ .Name | cyan }}",
			Selected: "{{ .Name | green }}",
		},
	}
	i, _, err := chaptersPrompt.Run()
	if err != nil {
		return err
	}

	contentPrompt := promptui.Select{
		Label: "Select a class to download",
		Items: fc.Chapters[i].Contents,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Active:   "❯ {{ .Title | magenta }}",
			Inactive: "   {{ .Title | cyan }}",
			Selected: "{{ .Title | green }}",
		},
	}
	j, _, err := contentPrompt.Run()
	if err != nil {
		return err
	}

	f, err := m3u8.DownloadVideo(
		fc.Chapters[i].Contents[j].VideoID,
		fc.Chapters[i].Contents[j].Title,
	)
	if err != nil {
		return err
	}

	ffmpeg := converter.NewFfmpeg()
	fmt.Println("Converting TS to MP4...")
	mp4, err := ffmpeg.TsToMP4(f)
	if err != nil {
		return nil
	}

	switch true {
	case upload.Gdrive != nil:
		moduleFolderID, err := upload.Gdrive.CreateFolder(
			fc.Modules[moduleIdx].Name,
			upload.RootFolderID,
		)
		if err != nil {
			return err
		}

		chapterFolderID, err := upload.Gdrive.CreateFolder(
			fc.Chapters[i].Name,
			moduleFolderID,
		)
		if err != nil {
			return err
		}

		fmt.Println("Uploading file to google drive...")
		_, err = upload.Gdrive.UploadMP4(fc.Chapters[i].Contents[j].Title, chapterFolderID, mp4)
		if err != nil {
			return err
		}
	default:
		fmt.Println("Creating local file...")
		dirPath = filepath.Join(dirPath, fc.Chapters[i].Name)
		if err = os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}

		filePath := filepath.Join(dirPath, fc.Chapters[i].Contents[j].Title+".mp4")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, mp4)
		return err
	}

	return nil
}
