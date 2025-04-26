package main

import "download-upload/cmd"

func main() {
	cmd.Execute()
}

// package main

// import (
// 	"fmt"

// 	"github.com/spf13/cobra"
// )

// func main() {
// 	var rootCmd = &cobra.Command{
// 		Use:   "app",
// 		Short: "An example CLI application",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			fmt.Println("Hello, World!")
// 		},
// 	}
// 	rootCmd.Execute()
// 	// fullCycle, err := streamings.FullCycle("josineyjunior14@gmail.com", "Jr11251423go?")
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }

// 	// fmt.Println(fullCycle.BearerToken)
// }
