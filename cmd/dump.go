/*
Copyright © 2025 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wolfmanjm/svd_lookup/internal"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the SVD database",
	Long: `Dump out the entire database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.Dump()
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
