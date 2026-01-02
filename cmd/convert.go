/*
Copyright © 2026 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wolfmanjm/svd_lookup/svd2db"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a .SVD file to a database file",
	Long: `Converts any .svd file to a database file for use with svn_lookup
	lookup_svd convert path/to/svd/file [outputfilename.db]
	`,
	Args: cobra.RangeArgs(1,2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ofile := ""
		if len(args) == 2 {
			ofile = args[1]
		}
		return svd2db.Convert(args[0], ofile)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
