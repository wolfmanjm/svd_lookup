/*
Copyright © 2026 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

// forthCmd represents the forth command
var forthCmd = &cobra.Command{
	Use:   "forth",
	Short: "Generate forth words to access the specified peripheral",
	Long: ``,
	Aliases: []string{"fc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.GenForthConsts(periph, reg_pat)
	},
}

func init() {
	forthCmd.Flags().StringVarP(&periph, "peripheral", "p", "", "Peripheral to use")
	forthCmd.Flags().StringVarP(&reg_pat, "register", "r", "", "Register pattern to filter on")

	if err := forthCmd.MarkFlagRequired("peripheral"); err != nil { panic(err) }
	rootCmd.AddCommand(forthCmd)
}
