/*
Copyright © 2025 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

var reg_pat string

// displayCmd represents the display command
var displayCmd = &cobra.Command{
	Use:   "display",
	Short: "Human readable display of the registers and fields for the specified peripheral",
	Long: `Human readable display of the registers and fields for the specified peripheral
	If -v is specified then descriptions for the registers and fields is also displayed
	If -r is specified then only the registers that match that pattern will be displayed
	The -p name may contain % as a wildcard for matching the peripheral name`,
	Aliases: []string{"d", "disp"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.Display(periph, reg_pat)
	},
}

func init() {
	displayCmd.Flags().StringVarP(&periph, "peripheral", "p", "", "Peripheral to use")
	displayCmd.Flags().StringVarP(&reg_pat, "register", "r", "", "Register pattern to filter on")

	if err := displayCmd.MarkFlagRequired("peripheral"); err != nil { panic(err) }
	rootCmd.AddCommand(displayCmd)
}
