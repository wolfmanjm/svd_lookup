package svd_lookup

import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"path/filepath"
	"errors"
	"strings"
	_ "sort"
)

/*
CREATE TABLE `mpus` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` varchar(255) NOT NULL UNIQUE, `description` varchar(255));
CREATE TABLE sqlite_sequence(name,seq);
CREATE TABLE `peripherals` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `mpu_id` integer, `derived_from_id` integer, `name` varchar(255) NOT NULL UNIQUE, `base_address` varchar(255), `description` varchar(255));
CREATE TABLE `registers` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `peripheral_id` integer, `name` varchar(255) NOT NULL, `address_offset` varchar(255), `reset_value` varchar(255), `description` varchar(255));
CREATE TABLE `fields` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `register_id` integer, `name` varchar(255) NOT NULL, `num_bits` integer, `bit_offset` integer, `description` varchar(255));
*/

type MPU struct {
	id int
	name string
	description string
}

type Peripheral struct {
	id int
	derived_from sql.Null[int]
	name string
	base_address string
	description sql.Null[string]
	registers *[]Register
}

type Register struct {
	id int
	name string
	address_offset string
	reset_value sql.Null[string]
	description sql.Null[string]
	fields *[]Field
}

type Field struct {
	id int
	name string
	num_bits int
	bit_offset int
	description sql.Null[string]
}

// print helpers for the structs
func (f Field) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Field:  %v, number bits %v, bit offset: %v\n", f.name, f.num_bits, f.bit_offset)
	return b.String()
}

func (r Register) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Register: %v, address offset: %v\n", r.name, r.address_offset)
	if r.fields != nil {
		for _, f := range *r.fields  {
			fmt.Fprint(&b, "    ", f)
		}
	}

	return b.String()
}

func (p Peripheral) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Peripheral: %v, base address: %v\n", p.name, p.base_address)
	if p.registers != nil {
		for _, r := range *p.registers  {
			fmt.Fprint(&b, "  ", r)
		}
	}
	return b.String()
}

var DB *sql.DB
var cwd string
var database string
var verbose bool

func FindUpwards(filename string) (string, error) {
	if cwd == "" {
		c, _ := os.Getwd()
		cwd= c
	}
	return lookupFile(cwd, filename)
}

func lookupFile(basepath string, filename string) (string, error) {
	matches, err := filepath.Glob(path.Join(basepath, filename))
	if len(matches) == 0 {
		return lookupInNearestDir(basepath, filename)
	}
	return matches[0], err
}

func lookupInNearestDir(basepath string, filename string) (string, error) {
	if basepath == "/" {
		return "", errors.New("database file " + filename + " not found. Starting at: " + cwd)
	}
	nearest := path.Dir(basepath)
	return lookupFile(nearest, filename)
}


func SetSearchPath(dir string) {
	cwd = dir
}

func SetDatabase(fn string) {
	database = fn
}

func SetVerbose() {
	verbose = true
}

func OpenDatabase() (error) {
	var dbfn string

	if database == "" {
		fn, err := FindUpwards("default-svd.db")
		if err != nil {
			return err
		}
		dbfn = fn
	} else {
		dbfn = database
	}

	// make sure database file exists
	_, err := os.Stat(dbfn)
	if err != nil {
		return fmt.Errorf("database file %v does not exist - %w", dbfn, err)
	}

	db, err := sql.Open("sqlite3", dbfn)
	if err != nil {
		return fmt.Errorf("Unable to open database file %v - %w", dbfn, err)
	}

	DB= db

	return nil
}

func CloseDatabase() {
	DB.Close()
}

func getMPU() string {
	var mpuname string
    // Query for a value based on a single row.
    if err := DB.QueryRow("SELECT name from mpus").Scan(&mpuname); err != nil {
        log.Fatal(err, " - Check this is a svd database")
    }

	return mpuname
}

func IntPow(base, exp int) int {
    result := 1
    for {
        if exp & 1 == 1 {
            result *= base
        }
        exp >>= 1
        if exp == 0 {
            break
        }
        base *= base
    }

    return result
}

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
		for _, r := range *pr.registers  {
			// filter out registers if required
			if reg_pat != "" && !strings.Contains(strings.ToLower(r.name), strings.ToLower(reg_pat)) {
				continue
			}

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

// collect all the registers and their fields for the named peripheral
func collect_registers(periph string) (Peripheral, error) {
	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		return p, fmt.Errorf("Peripheral %v not found: %w", periph, err)
	}

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
	}

	regs, err := fetch_registers(id)
	if err != nil {
		return p, err
	}
	for i, r := range regs {
		fields, err := fetch_fields(r.id)
		if err != nil {
			return p, err
		}
		regs[i].fields = &fields
	}

	p.registers = &regs

	return p, nil
}

func Dump() (error) {
	fmt.Println("MPU: ", getMPU())
	fmt.Println("Database Dump:")

	periphs, err := fetch_peripherals()

	if err != nil {
		return err
	}

	for _, p := range periphs {
		if p.derived_from.Valid {
			fmt.Print(p)
			p, err := fetch_peripheral(p.derived_from.V)
			if err != nil {
				return fmt.Errorf("No derived peripheral with id: %v found: %w", p.derived_from.V, err)
			}else{
				fmt.Println("  Registers the same as ", p.name)
			}

		} else {
			// registers := fetch_registers(p_id)
			// dump_registers_and_fields(registers)
			pr, err := collect_registers(p.name)
			if err != nil {
				return err
			}
			fmt.Print(pr)
		}
		fmt.Println()
	}

	return nil
}

func List() (error) {
	fmt.Println("Available Peripherals for MPU: ", getMPU())

	periphs, err := fetch_peripherals()
	if err != nil {
		return err
	}

	for _, p := range periphs {
		if verbose && p.description.Valid {
			fmt.Println(p.name, " - ", p.description.V)
		} else {
			fmt.Println(p.name)
		}
	}

	return nil
}

func Registers(periph string) (error) {
	fmt.Println("Registers for Peripheral: ", periph, " for MPU: ", getMPU())

	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		fmt.Println("  No peripheral with name like: ", periph)
		return err
	}

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
	}

	regs, err := fetch_registers(id)
	if err != nil {
		return err
	}
	for _, r := range regs {
		fmt.Print(r.name)
		if verbose && r.description.Valid {
			fmt.Print(" - ", r.description.V)
		}
		fmt.Println()
	}

	return nil
}

// database helpers

func fetch_peripherals() ([]Peripheral, error) {
	var periphs []Peripheral
	periph_rows, err := DB.Query("select id, derived_from_id, name, base_address, description from peripherals ORDER BY name")
	if err != nil {
		return periphs, err
	}
	defer periph_rows.Close()

	for periph_rows.Next() {
		var p Peripheral
		err = periph_rows.Scan(&p.id, &p.derived_from, &p.name, &p.base_address, &p.description)
		if err != nil {
			return periphs, err
		}
		periphs= append(periphs, p)
	}

	err = periph_rows.Err()
	if err != nil {
		return periphs, err
	}

	return periphs, nil
}

func fetch_peripheral_by_name(periph string) (Peripheral, error) {
	var p Peripheral

    if err := DB.QueryRow("SELECT id, derived_from_id, name, base_address, description from peripherals WHERE lower(name) LIKE lower(?)", periph).
    	Scan(&p.id, &p.derived_from, &p.name, &p.base_address, &p.description); err != nil {
        	return p, err
    }
    return p, nil;
}

func fetch_peripheral(id int) (Peripheral, error) {
	var p Peripheral

    if err := DB.QueryRow("SELECT id, derived_from_id, name, base_address, description from peripherals WHERE id = ?", id).
    	Scan(&p.id, &p.derived_from, &p.name, &p.base_address, &p.description); err != nil {
        	return p, err
    }
    return p, nil;
}

func fetch_registers(p_id int) ([]Register, error) {
	register_rows, err := DB.Query("select id, name, address_offset, reset_value, description from registers WHERE peripheral_id = ? ORDER BY name", p_id)

	if err != nil {
		return nil, fmt.Errorf("failure in fetch_registers query for id %v: %w", p_id, err)
	}
	defer register_rows.Close()

	var registers []Register
	for register_rows.Next() {
		var reg Register
		err = register_rows.Scan(&reg.id, &reg.name, &reg.address_offset, &reg.reset_value, &reg.description)
		if err != nil {
			return nil, fmt.Errorf("failure in fetch_registers scan for id %v: %w", p_id, err)
		}
		registers= append(registers, reg)
	}

	err = register_rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failure in fetch_registers rows for id %v: %w", p_id, err)
	}

	return registers, nil
}

func fetch_fields(r_id int) ([]Field, error) {
	field_rows, err := DB.Query("select name, num_bits, bit_offset, description from fields WHERE register_id = ? ORDER BY bit_offset", r_id)

	if err != nil {
		return nil, fmt.Errorf("failure in fetch_fields query for id %v: %w", r_id, err)
	}
	defer field_rows.Close()
	var fields []Field
	for field_rows.Next() {
		var f Field
		err = field_rows.Scan(&f.name, &f.num_bits, &f.bit_offset, &f.description)
		if err != nil {
			return nil, fmt.Errorf("failure in fetch_fields scan for id %v: %w", r_id, err)
		}
		fields= append(fields, f)
	}

	err = field_rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failure in fetch_registers rows for id %v: %w", r_id, err)
	}

	return fields, nil
}
