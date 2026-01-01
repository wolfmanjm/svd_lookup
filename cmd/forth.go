/*
Copyright © 2026 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

var forth_type bool
// forthCmd represents the forth command
var forthCmd = &cobra.Command{
	Use:   "forth",
	Short: "Generate forth words to access the specified peripheral",
	Long: `Generates forth words to access the specified peripheral
	By default it generates constants, by using the --freg flag it will instead generate words that use the register format`,
	Aliases: []string{"fth"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if forth_type {
			return svd_lookup.GenForthRegs(periph, reg_pat)
		} else {
			return svd_lookup.GenForthConsts(periph, reg_pat)
		}
	},
}

func init() {
	forthCmd.Flags().StringVarP(&periph, "peripheral", "p", "", "Peripheral to use")
	forthCmd.Flags().StringVarP(&reg_pat, "register", "r", "", "Register pattern to filter on")
	forthCmd.Flags().BoolVar(&forth_type, "freg", false, "Generate register format")

	if err := forthCmd.MarkFlagRequired("peripheral"); err != nil { panic(err) }
	rootCmd.AddCommand(forthCmd)
}
