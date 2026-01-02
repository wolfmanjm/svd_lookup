package svd2db

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"path/filepath"
)

// SVD structure definitions based on CMSIS-SVD specification
type Device struct {
	XMLName     xml.Name     `xml:"device"`
	Name        string       `xml:"name"`
	Description string       `xml:"description"`
	Peripherals []Peripheral `xml:"peripherals>peripheral"`
}

type Peripheral struct {
	Name         string       `xml:"name"`
	Description  string       `xml:"description"`
	BaseAddress  string       `xml:"baseAddress"`
	GroupName    string       `xml:"groupName"`
	Registers    []Register   `xml:"registers>register"`
	DerivedFrom  string       `xml:"derivedFrom,attr"`
	AddressBlock AddressBlock `xml:"addressBlock"`
}

type AddressBlock struct {
	Offset string `xml:"offset"`
	Size   string `xml:"size"`
	Usage  string `xml:"usage"`
}

type Register struct {
	Name        string  `xml:"name"`
	Description string  `xml:"description"`
	Offset      string  `xml:"addressOffset"`
	Size        string  `xml:"size"`
	Access      string  `xml:"access"`
	ResetValue  string  `xml:"resetValue"`
	Fields      []Field `xml:"fields>field"`
}

type Field struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	BitOffset   string `xml:"bitOffset"`
	BitWidth    string `xml:"bitWidth"`
	BitRange    string `xml:"bitRange"`
	Access      string `xml:"access"`
	LSB         string `xml:"lsb"`
	MSB         string `xml:"msb"`
}

// keeps a list of peripherals to id mapping, needed for derived_from peripherals
var periph_ids map[string]int
// var deferred_derived_from map[string]string

func Convert(filename string, ofile string) error {
	// Read the SVD file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Error in convert reading file: %w\n", err)
	}

	// Parse XML
	var device Device
	err = xml.Unmarshal(data, &device)
	if err != nil {
		return fmt.Errorf("Error in convert parsing XML: %w\n", err)
	}

	var outfile string
	if ofile == "" {
		outfile = strings.Replace(filename, filepath.Ext(filename), ".db", 1)
	} else {
		outfile = ofile
	}

	// create the database with its schema
	db, err := db_createdb(outfile)
	if err != nil {
		return fmt.Errorf("Error in convert creating database: %w\n", err)
	}
	defer db.Close()

	// Add device information to database
	m := map[string]any{"name": device.Name, "description": device.Description}
	mpu_id, err := db_insert(db, "mpus", m)
	if err != nil {
		return fmt.Errorf("Error in convert inserting 'mpu' to database: %w\n", err)
	}

	periph_ids = make(map[string]int)
	// deferred_derived_from = make(map[string]string)

	// insert peripherals and their registers
	for _, peripheral := range device.Peripherals {
		if err := insertPeripheral(db, mpu_id, peripheral); err != nil {
			return fmt.Errorf("Error in convert inserting 'peripherals' to database: %w\n", err)
		}
	}

	// now process any derived from as needed

	return nil
}

func insertPeripheral(db *sql.DB, mpu_id int, p Peripheral) error {
	fmt.Println("Processing Peripheral: " + p.Name)

	m := map[string]any{"name": p.Name, "mpu_id": mpu_id, "base_address": p.BaseAddress}

	if p.Description != "" {
		m["description"] = p.Description
	}

	derived_from_flg := false
	if p.DerivedFrom != "" {
		elem, ok := periph_ids[p.DerivedFrom]
		if ok {
			m["derived_from_id"] = elem
		} else {
			// put it on the list to enter when we have processed all the other peripherals
			// deferred_derived_from[p.Name] = p.DerivedFrom
			return fmt.Errorf("ERROR peripheral %v derived from %v not yet entered\n", p.Name, p.DerivedFrom)
		}
		derived_from_flg = true
	}

	// enter into the database
	peripheral_id, err := db_insert(db, "peripherals", m)
	if err != nil {
		return fmt.Errorf("Error in insertPeripheral inserting peripheral %v to database: %w\n", p.Name, err)
	}

	// add to list of peripherals and the ids
	periph_ids[p.Name] = peripheral_id

	if !derived_from_flg && len(p.Registers) > 0 {
		// Insert registers
		for _, register := range p.Registers {
			if err := insertRegister(db, peripheral_id, register); err != nil {
				return fmt.Errorf("Error in insertPeripheral inserting registers to database: %w\n",  err)
			}
		}
	}

	return nil
}

func insertRegister(db *sql.DB, peripheral_id int, r Register) error {
	// fmt.Println("Processing Register: " + r.Name)
	m := map[string]any{"name": r.Name, "peripheral_id": peripheral_id, "address_offset": r.Offset}

	if r.Description != "" {
		m["description"] = r.Description
	}

	if r.ResetValue != "" {
		m["reset_value"] = r.ResetValue
	}

	// enter into the database
	register_id, err := db_insert(db, "registers", m)
	if err != nil {
		return fmt.Errorf("Error in insertRegister inserting %v to database: %w\n", r.Name, err)
	}

	// Insert fields
	if len(r.Fields) > 0 {
		for _, field := range r.Fields {
			if err := insertField(db, register_id, field); err != nil {
				return fmt.Errorf("Error in insertRegister inserting fields to database: %w\n",  err)
			}
		}
	}

	return nil
}

func insertField(db *sql.DB, register_id int, f Field) error {
	// fmt.Println("Processing Field: " + f.Name)
	m := map[string]any{"name": f.Name, "register_id": register_id}

	if f.Description != "" {
		m["description"] = f.Description
	}

	// Handle different bit position formats and convert to num_bits and bit_offset
	var bit_offset, num_bits int
	var err error
	if f.BitOffset != "" && f.BitWidth != "" {
		bit_offset, err = strconv.Atoi(f.BitOffset)
		if err != nil {
			return fmt.Errorf("Error in insertField converting bitOffset %v to integer: %w\n", f.BitOffset, err)
		}
		num_bits, err = strconv.Atoi(f.BitWidth)
		if err != nil {
			return fmt.Errorf("Error in insertField converting bitWidth %v to integer: %w\n", f.BitWidth, err)
		}

	} else if f.BitRange != "" {
		// Bit Range: [31:16]
		br := strings.Split(f.BitRange[1:len(f.BitRange)-1], ":")
		hr, err := strconv.Atoi(br[0])
		if err != nil {
			return fmt.Errorf("Error in insertField converting bitRange to integer: %w\n",  err)
		}
		lr, err := strconv.Atoi(br[1])
		if err != nil {
			return fmt.Errorf("Error in insertField converting bitRange to integer: %w\n",  err)
		}
		bit_offset = lr
		num_bits = (hr-lr)+1

	} else if f.LSB != "" && f.MSB != "" {
		return fmt.Errorf("Error in insertField bit MSB/LSB not handled")

	} else {
		return fmt.Errorf("Error in insertField no valid bit info found")
	}

	m["num_bits"] = num_bits
	m["bit_offset"] =  bit_offset

	// enter into the database
	_, err = db_insert(db, "fields", m)
	if err != nil {
		return fmt.Errorf("Error in insertField inserting %v to database: %w\n", f.Name, err)
	}

	return nil
}
