package svd_lookup

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// code snippets for forth helpers
var modify_reg_code string = `
: modify-reg ( value mask pos reg -- )
    >r tuck         \ -- value pos mask pos
    lshift r@ bic!  \ clear mask first
    lshift r> bis!  \ set the value bits
;
`
var lib_registers_code string = `
: registers ( -- )
    0 ;             \ offset start

: reg
    <builds         ( offset -- newoffset )
        dup , cell+
    does>           ( structure-base -- structure-member-address )
        @ + ;

: regC
    <builds         ( offset -- newoffset )
        dup , cell+
    does>           ( structure-base stream -- structure-member-address )
        @ swap $18 * + + ;

: end-registers ( -- )
    drop ;          \ last offset

\ bit masks
: bit ( n -- n )
    1 swap lshift 1-foldable ;
`

var Addwords bool

// generate forth constants for the specified peripheral
func GenForthConsts(periph string, reg_pat string) error {
    // collects and populates all the registers and fields for this peripheral
    pr, err := collect_registers(periph)
    if err != nil {
        return fmt.Errorf("Failed to collect registers for peripheral %v: %w", periph, err)
    }

    if Addwords {
        fmt.Print(modify_reg_code)
        fmt.Println()
    }

    fmt.Printf("%v constant %v_BASE\n", strings.Replace(pr.base_address, "0x", "$", 1), pr.name)

    prefix := strings.ToLower(pr.name)[0:3]

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
        for _, r := range regs {
            fmt.Printf("  \\ Bitfields for %v_%v\n", prefix, r.name)

            // create constants for the bit fields
            // m_ use with modify-reg ( value mask pos reg -- )
            // ie 5 m_CR2_TSER SPI1 _spCR2 modify-reg
            // b_ use either bic! or bis!
            // ie b_CR1_SSI SPI2 _sCR1 bis!
            if r.fields != nil {
                for _, f := range *r.fields {
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

func GenForthRegs(periph string, reg_pat string) error {
    // collects and populates all the registers and fields for this peripheral
    pr, err := collect_registers(periph)
    if err != nil {
        return fmt.Errorf("Failed to collect registers for peripheral %v: %w", periph, err)
    }

    if Addwords {
        fmt.Print(lib_registers_code)
        fmt.Println()
    }

    fmt.Printf("%v constant %v\n", strings.Replace(pr.base_address, "0x", "$", 1), pr.name)

    fmt.Println("  registers")
    prefix := strings.ToLower(pr.name)[0:2]
    addr := 0

    // print out
    if pr.registers != nil {
        regs := *pr.registers
        if reg_pat != "" {
            // filter out registers if required
            regs = slices.DeleteFunc(regs, func(n Register) bool {
                return !strings.Contains(strings.ToLower(n.name), strings.ToLower(reg_pat))
            })
        }

        // sort by address_offset (which is a string)
        sort.Slice(regs, func(i, j int) bool {
            a, _ := strconv.ParseUint(regs[i].address_offset[2:], 16, 32)
            b, _ := strconv.ParseUint(regs[j].address_offset[2:], 16, 32)
            return a < b
        })

        // print out register constants
        for _, r := range regs {
            a, err := strconv.ParseUint(r.address_offset[2:], 16, 32)
            if err != nil {
                return fmt.Errorf("Unable to parse hex %v - %w", r.address_offset, err)
            }

            if int(a) != addr {
                fmt.Printf("    drop $%08X\n", a)
                addr = int(a)
            }
            addr += 4
            fmt.Printf("    reg _%v%v\n", prefix, r.name)
        }
        fmt.Println("  end-registers")

        // print out the fields for each register
        for _, r := range regs {
            fmt.Printf("\n\\ Bitfields for %v\n", r.name)
            if r.fields != nil {
                for _, f := range *r.fields {
                    bf := r.name + "_" + f.name
                    if f.num_bits == 1 {
                        fmt.Printf("  %v bit constant b_%v\n", f.bit_offset, bf)
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
