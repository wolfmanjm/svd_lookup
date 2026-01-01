/*
Copyright © 2025 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
	"github.com/spf13/cobra"
)

// registersCmd represents the registers command
var registersCmd = &cobra.Command{
	Use:   "registers",
	Short: "List all the registers for the specified peripheral",
	Long: `Just a list of register names`,
	Aliases: []string{"r", "regs"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.Registers(periph)
	},
}

func init() {
	registersCmd.Flags().StringVarP(&periph, "peripheral", "p", "", "Peripheral to use")
	if err := registersCmd.MarkFlagRequired("peripheral"); err != nil { panic(err) }
	rootCmd.AddCommand(registersCmd)
}
