package cmd

import (
	s "download-upload/cmd/streamings"
	"download-upload/pkg/streamings"
	"fmt"
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("===============================================================")
		fmt.Println(" _____ _                            _                   ")
		fmt.Println("/  ___| |                          (_)                  ")
		fmt.Println("\\ `--.| |_ _ __ ___  __ _ _ __ ___  _ _ __   __ _       ")
		fmt.Println(" `--. \\ __| '__/ _ \\/ _` | '_ ` _ \\| | '_ \\ / _` |      ")
		fmt.Println("/\\__/ / |_| | |  __/ (_| | | | | | | | | | | (_| |      ")
		fmt.Println("\\____/ \\__|_|  \\___|\\__,_|_| |_| |_|_|_| |_|\\__, |      ")
		fmt.Println("                                             __/ |      ")
		fmt.Println("                                            |___/       ")
		fmt.Println("______                    _                 _           ")
		fmt.Println("|  _  \\                  | |               | |          ")
		fmt.Println("| | | |_____      ___ __ | | ___   __ _  __| | ___ _ __ ")
		fmt.Println("| | | / _ \\ \\ /\\ / / '_ \\| |/ _ \\ / _` |/ _` |/ _ \\ '__|")
		fmt.Println("| |/ / (_) \\ V  V /| | | | | (_) | (_| | (_| |  __/ |   ")
		fmt.Println("|___/ \\___/ \\_/\\_/ |_| |_|_|\\___/ \\__,_|\\__,_|\\___|_| ")
		fmt.Print("===============================================================\n")
		fmt.Println("\nAvailable sources:")

		prompt := promptui.Select{
			Label: "Select a source to download from",
			Items: streamings.Sources,
		}

		_, selectedSource, err := prompt.Run()
		if err != nil {
			fmt.Printf("Selection failed %v\n", err)
			return
		}

		switch selectedSource {
		case string(streamings.FullCycleSrc):
			err = s.FullCycleCmd.RunE(s.FullCycleCmd, args)
			if err != nil {
				log.Fatal("error on run cmd", err)
			}
		}

	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(s.FullCycleCmd)
}
