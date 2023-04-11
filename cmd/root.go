/*
Copyright © 2023 Arpan Adhikari

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ktm",
	Short: "A Kubernetes Time Machine",
	Long:  `KTM is a tool that records your kubernetes cluster and allows you to play it back in real-time.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
