package svd_lookup

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

func Display(periph string, reg_pat string) (error) {
	fmt.Println("Registers and fields for Peripheral:", periph, " for MPU:", getMPU())

	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		return errors.New("No peripheral with name like: " + periph)
	}

	fmt.Printf("%v base address: %v\n", p.name, p.base_address);

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
		np, _ := fetch_peripheral(id)
		fmt.Println("Has the same registers as", np.name)
	}

	// collects and populates all the registers and fields for this peripheral
	pr, err := collect_registers(p.name)
	if err != nil {
		return fmt.Errorf("Failed to collect registers for peripheral %v: %w", p.name, err)
	}

	// print out
	if pr.registers != nil {
		regs := *pr.registers

		if reg_pat != "" {
			// filter out registers if required
			regs = slices.DeleteFunc(regs, func(n Register) bool {
				return !strings.Contains(strings.ToLower(n.name), strings.ToLower(reg_pat))
			})
		}

		for _, r := range regs {
			s := fmt.Sprintf("Register %v offset: %v, reset: %v", r.name, r.address_offset, r.reset_value.V)
			if verbose && r.description.Valid {
				s += " - " + r.description.V
			}
			fmt.Println(s)

			// print out the fields for this register
			if r.fields != nil {
				for _, f := range *r.fields  {
					desc := ""
					if verbose && f.description.Valid {
						desc = " - " + f.description.V
					}
					mask := (IntPow(2, f.num_bits) - 1) << f.bit_offset
					fmt.Printf("    %v: number bits %v, bit offset: %v, mask: 0x%08X %s\n", f.name, f.num_bits, f.bit_offset, mask, desc)
				}
			}
		}
	}

	return nil
}

