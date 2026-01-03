/*
Copyright © 2026 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

// asmCmd represents the asm command
var asmCmd = &cobra.Command{
	Use:   "asm --peripheral name [--register regpattern]",
	Short: "Generate asm .equ directives defining register and fields",
	Long:  `Generate asm .equ directives defining register and fields
		specifying the -r 'pat' will only print out the registers with 'pat' in the name
		to print out just the register equates use -r xx
		if the periphal name ends in '_n' then we scan for all matching peripherals
		that end in a number and output them first
		eg SPI_n will get SPI0 SPI1 SPI2 etc or TIM_ will get TIM1 TIM2 ... TIM12 etc
		in this case the register and fields will be generic to any of the registers printed out
		`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return svd_lookup.GenAsm(periph, reg_pat)
	},
}

func init() {
	asmCmd.Flags().StringVarP(&periph, "peripheral", "p", "", "Peripheral to use")
	asmCmd.Flags().StringVarP(&reg_pat, "register", "r", "", "Register pattern to filter on")
	if err := asmCmd.MarkFlagRequired("peripheral"); err != nil { panic(err) }

	rootCmd.AddCommand(asmCmd)
}
