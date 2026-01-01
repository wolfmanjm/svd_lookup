/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all peripherals",
	Long: `List the Peripheral names available`,
	Aliases: []string{"l", "lst"},
	Run: func(cmd *cobra.Command, args []string) {
		svd_lookup.List()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

}
