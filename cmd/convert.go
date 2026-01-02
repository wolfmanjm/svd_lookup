/*
Copyright © 2026 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wolfmanjm/svd_lookup/svd2db"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert file.svd [outfile.db]",
	Short: "Convert a .SVD file to a database file",
	Long: `Converts any .svd file to a database file for use with svn_lookup
	No flags are required and the output filename is optional
	`,
	Args: cobra.RangeArgs(1,2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cwd != "" || database != "" {
			return fmt.Errorf("the -c and -d flags are ignored and should not be used with convert")
		}
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
