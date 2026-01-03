package svd_lookup

import (
	"fmt"
	"slices"
	"strings"
)

// generate assembly defines for the specified peripheral
func GenAsm(periph string, reg_pat string) error {
    // collects and populates all the registers and fields for this peripheral
    pr, err := collect_registers(periph)
    if err != nil {
        return fmt.Errorf("Failed to collect registers for peripheral %v: %w", periph, err)
    }

    fmt.Printf(".equ %v_BASE, %v\n", pr.name, pr.base_address)

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
            fmt.Printf("  .equ _%v, %v\n", r.name, r.address_offset)
        }

        // print out the fields for each register
        for _, r := range regs {
            fmt.Printf("; Bitfields for _%v\n", r.name)
            if r.fields != nil {
                for _, f := range *r.fields {
                    bf := r.name + "_" + f.name
                    if f.num_bits == 1 {
                        fmt.Printf("  .equ b_%v, 1<<%v\n", bf, f.bit_offset)
                    } else {
                        mask := (IntPow(2, f.num_bits) - 1) << f.bit_offset
                        fmt.Printf("  .equ m_%v, 0x%08X\n", bf, mask)
                        fmt.Printf("  .equ o_%v, %v\n", bf, f.bit_offset)
                    }
                }
            }
        }
    }

    return nil
}
