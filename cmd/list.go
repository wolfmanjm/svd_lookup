/*
Copyright © 2025 Jim Morris <morris@wolfman.com>
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.List()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

}
