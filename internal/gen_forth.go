package svd_lookup

import (
	"fmt"
	"slices"
	"strings"
)

// generate forth constants for the specified peripheral
func GenForthConsts(periph string, reg_pat string) (error) {
	// collects and populates all the registers and fields for this peripheral
	pr, err := collect_registers(periph)
	if err != nil {
		return fmt.Errorf("Failed to collect registers for peripheral %v: %w", periph, err)
	}

	fmt.Printf("%v constant %v_BASE\n", strings.Replace(pr.base_address, "0x", "$", 1), pr.name);

	prefix := strings.ToLower(pr.name)[0:2]

	// print out
	if pr.registers != nil {
		regs := *pr.registers
		if reg_pat != "" {
			// filter out registers if required
			regs = slices.DeleteFunc(regs, func(n Register) bool {
				return !strings.Contains(strings.ToLower(n.name), strings.ToLower(reg_pat))
			})
		}

		// print out register constants
		for _, r := range regs {
			a := strings.Replace(r.address_offset, "0x", "$", 1)
			fmt.Printf("  %v_BASE %v + constant %v_%v\n", pr.name, a, prefix, r.name)
		}

		// print out the fields for each register
		for _, r := range regs  {
			fmt.Printf("  \\ Bitfields for %v_%v\n", prefix, r.name)

			// create constants for the bit fields
			// m_ use with modify-reg ( value mask pos reg -- )
			// ie 5 m_CR2_TSER SPI1 _spCR2 modify-reg
			// b_ use either bic! or bis!
			// ie b_CR1_SSI SPI2 _sCR1 bis!
			if r.fields != nil {
				for _, f := range *r.fields  {
					bf := prefix + "_" + r.name + "_" + f.name
					if f.num_bits == 1 {
						fmt.Printf("  1 %v lshift constant b_%v\n", f.bit_offset, bf)
					} else {
						mask := (IntPow(2, f.num_bits) - 1)
						fmt.Printf("  $%08X %v 2constant m_%v\n", mask, f.bit_offset, bf)
					}
				}
			}
		}
	}

	return nil
}

func GenForthRegs(periph string, reg_pat string) (error) {
	fmt.Println("forth regs")
	return nil
}
