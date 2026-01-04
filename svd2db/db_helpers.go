package svd2db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	_ "modernc.org/sqlite"
)

/*
	SVD Database schema
	CREATE TABLE `mpus` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` varchar(255) NOT NULL UNIQUE, `description` varchar(255));

	CREATE TABLE sqlite_sequence(name,seq);

	CREATE TABLE `peripherals` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `mpu_id` integer, `derived_from_id` integer, `name` varchar(255) NOT NULL UNIQUE, `base_address` varchar(255), `description` varchar(255));

	CREATE TABLE `registers` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `peripheral_id` integer, `name` varchar(255) NOT NULL, `address_offset` varchar(255), `reset_value` varchar(255), `description` varchar(255));

	CREATE TABLE `fields` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `register_id` integer, `name` varchar(255) NOT NULL, `num_bits` integer, `bit_offset` integer, `description` varchar(255));
*/

func db_createdb(filename string) (*sql.DB, error) {

	// make sure database file does not exist yet
	_, err := os.Stat(filename)
	if err == nil {
		return nil, fmt.Errorf("database file %v already exists", filename)
	}

	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open database file %v - %w", filename, err)
	}

	sqlStmt := `
CREATE TABLE mpus (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, name text NOT NULL UNIQUE, description text);
CREATE TABLE peripherals (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, mpu_id integer NOT NULL, derived_from_id integer, name text NOT NULL UNIQUE, base_address text NOT NULL, description text);
CREATE TABLE registers (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, peripheral_id integer NOT NULL, name text NOT NULL, address_offset text NOT NULL, reset_value text, description text);
CREATE TABLE fields (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, register_id integer NOT NULL, name text NOT NULL, num_bits integer NOT NULL, bit_offset integer NOT NULL, description text);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return  nil, fmt.Errorf("Unable to create tables: %v: %v\n", err, sqlStmt)
	}

	return db, nil
}

func db_insert(db *sql.DB, table string, data map[string]any) (int, error) {
	n := len(data)
	var keys []string
	var values []any

	for k, v := range data {
		keys = append(keys, k)
		values = append(values, v)
	}

	s := " (" + strings.Join(keys, ", ") + ") "

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("db_insert begin failed: %w", err)
	}

	st := "insert into " + table + s + "values(" + strings.Repeat("?,", n-1) + "?)"
	stmt, err := tx.Prepare(st)
	if err != nil {
		return 0, fmt.Errorf("db_insert prepare failed: %w", err)
	}
	defer stmt.Close()

	r, err := stmt.Exec(values...)
	if err != nil {
		return 0, fmt.Errorf("db_insert exec failed: %w", err)
	}

	idx, err:= r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db_insert lastinsertid failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("db_insert commit failed: %w", err)
	}
	return int(idx), nil
}
