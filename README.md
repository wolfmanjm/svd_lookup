This project is a rewrite of my svd [lookup tools](https://github.com/wolfmanjm/svd2db-v2) that were written in ruby ported to GO.
There is also program that converts a .svd to a sqlite3 database.

This program is a commandline program written in GO (as a way to learn GO).

It currently allows querying the svd database and displaying things like all
peripherals, all registers in a periperal and a human readable display of the
registers and fields for a given peripheral.

It can also generate words for forth allowing access to the peripherals and
registers, one of two formats can be selected:

1. Use constants to define the registers and bitfields
2. use the lib_registers structs to access them

The helper words supporting these two methods can optionally be added to the
output with the --addwords option

`svd_lookup forth --help` gives more details

Additionally it can generate defines (.equ) for assembly level code (risc-v or arm)

`svd_lookup asm --help` gives more details

You can specify the database to use with the --database option, if this is not specified
then it will search in the current directory and above for a default-svd.db file and use that.
You can set the start directory to search from with the --curdir option.

The data directory has some example SVD databases already converted.

The bins/ directory has various binaries ready to run on selected platforms.

To convert a .SVD file to the database you would run...

```
svd_lookup convert myfile.svd myfile.db
```

This may take a while for very large SVD files

```
> svd_lookup --help
Query a SVD database in various ways.
Depending on the subcommand it can generate various code sequences to access the peripherals and registers
or display the available peripherals and/or registers in a human readable way.

Usage:
	svd_lookup [command]

Available Commands:
	asm         Generate asm .equ directives defining register and fields
	completion  Generate the autocompletion script for the specified shell
	convert     Convert a .SVD file to a database file
	display     Human readable display of the registers and fields for the specified peripheral
	dump        Dumps the SVD database
	forth       Generate forth words to access the specified peripheral
	help        Help about any command
	list        List all peripherals
	registers   List all the registers for the specified peripheral

Flags:
	-c, --curdir string     set the current directory for db search
	-d, --database string   use the named database
	-h, --help              help for svd_lookup
	-v, --verbose           verbose output

Use "svd_lookup [command] --help" for more information about a command.
```
