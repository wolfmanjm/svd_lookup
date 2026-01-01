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
		return "", errors.New("database file not found: " + filename + " Starting at: " + cwd)
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

	_, err := os.Stat(dbfn)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbfn)
	if err != nil {
		return err
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

// TODO this can be refactored
func Display(periph string, reg_pat string) {
	fmt.Println("Registers and fields for Peripheral:", periph, " for MPU:", getMPU())

	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		fmt.Println("  No peripheral with name like: ", periph, err)
		return
	}

	fmt.Printf("%v base address: %v\n", p.name, p.base_address);

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
		np, _ := fetch_peripheral(id)
		fmt.Println("Has the same registers as", np.name)
	}

	regs := fetch_registers(id)
	m := make(map[int]string)
	var v []int
	for _, r := range regs {
		if reg_pat != "" && !strings.Contains(strings.ToLower(r.name), strings.ToLower(reg_pat)) {
			continue
		}
		s := fmt.Sprintf("Register %v offset: %v, reset: %v", r.name, r.address_offset, r.reset_value.V)
		if verbose && r.description.Valid {
			s += " - " + r.description.V
		}
		m[r.id] = s
		v = append(v, r.id) // so we can keep the correct register order
	}

	// we want to display the registers in the order they were given us (alphabetical)
	for _, i := range v {
		s := m[i]
		fmt.Println(s)
		fields := fetch_fields(i)
		for _,f := range fields {
			desc := ""
			if verbose && f.description.Valid {
				desc = " - " + f.description.V
			}

			mask := (IntPow(2, f.num_bits) - 1) << f.bit_offset
			fmt.Printf("  %v: number bits %v, bit offset: %v, mask: 0x%08X %s\n", f.name, f.num_bits, f.bit_offset, mask, desc)
		}
		fmt.Println()
	}
}

// collect all the registers and their fields for the named peripheral
func collect_registers(periph string) Peripheral {
	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		fmt.Println("  No peripheral with name like: ", periph, err)
		return p
	}

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
	}

	regs := fetch_registers(id)

	for i, r := range regs {
		fields := fetch_fields(r.id)
		regs[i].fields = &fields
	}

	p.registers = &regs

	return p
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

func Dump() {
	fmt.Println("MPU: ", getMPU())
	fmt.Println("Database Dump:")

	periphs, err := fetch_peripherals()

	if err != nil {
		log.Fatal(err)
	}

	for _, p := range periphs {
		if p.derived_from.Valid {
			fmt.Print(p)
			p, err := fetch_peripheral(p.derived_from.V)
			if err != nil {
				fmt.Println("  No peripheral with id: ", p.derived_from.V, err)
			}else{
				fmt.Println("  Registers the same as ", p.name)
			}

		} else {
			// registers := fetch_registers(p_id)
			// dump_registers_and_fields(registers)
			pr := collect_registers(p.name)
			fmt.Print(pr)
		}
		fmt.Println()
	}
}

func List() {
	fmt.Println("Available Peripherals for MPU: ", getMPU())

	periphs, err := fetch_peripherals()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range periphs {
		if verbose && p.description.Valid {
			fmt.Println(p.name, " - ", p.description.V)
		} else {
			fmt.Println(p.name)
		}
	}
}

func Registers(periph string) {
	fmt.Println("Registers for Peripheral: ", periph, " for MPU: ", getMPU())

	p, err := fetch_peripheral_by_name(periph)
	if err != nil {
		fmt.Println("  No peripheral with name like: ", periph, err)
		return
	}

	id := p.id

	if p.derived_from.Valid {
		id = p.derived_from.V
	}

	regs := fetch_registers(id)

	for _, r := range regs {
		fmt.Print(r.name)
		if verbose && r.description.Valid {
			fmt.Print(" - ", r.description.V)
		}
		fmt.Println()
	}
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

func fetch_registers(p_id int) []Register {
	register_rows, err := DB.Query("select id, name, address_offset, reset_value, description from registers WHERE peripheral_id = ? ORDER BY name", p_id)
	defer register_rows.Close()

	if err != nil {
		log.Fatal(err)
	}
	var registers []Register
	for register_rows.Next() {
		var reg Register
		err = register_rows.Scan(&reg.id, &reg.name, &reg.address_offset, &reg.reset_value, &reg.description)
		if err != nil {
			log.Fatal(err)
		}
		registers= append(registers, reg)
	}

	err = register_rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return registers
}

func fetch_fields(r_id int) []Field {
	field_rows, err := DB.Query("select name, num_bits, bit_offset, description from fields WHERE register_id = ? ORDER BY bit_offset", r_id)
	defer field_rows.Close()

	if err != nil {
		log.Fatal(err)
	}
	var fields []Field
	for field_rows.Next() {
		var f Field
		err = field_rows.Scan(&f.name, &f.num_bits, &f.bit_offset, &f.description)
		if err != nil {
			log.Fatal(err)
		}
		fields= append(fields, f)
	}

	err = field_rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return fields
}
