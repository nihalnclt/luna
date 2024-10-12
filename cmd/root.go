package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	Reset = "\033[0m"
	Green = "\033[32m"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "luna",
	Short: "Luna short desc",
	Long:  "Luna long desc",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
