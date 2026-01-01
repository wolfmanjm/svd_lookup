This project is a rewrite of my svd [lookup tools](https://github.com/wolfmanjm/svd2db-v2) that were written in ruby ported to GO.
The program that converts a .svd to a sqlite3 database is still written in ruby in that other repo.

This program is a commandline program written in GO (as a way to learn GO).

It currently allows querying the svd database and displaying things like all peripherals, all registers in a periperal and a human readable
display of the registers and fields for a given peripheral

The data directory has some example SVD databases already converted.

The bins/ directory has various binaries ready to run on selected platforms.

```
> svd_lookup --help
access the SVD database,
and depending on the subcommand generate various code sequences to access the peripherals and registers
or display the available peripherals and/or registers.

Usage:
svd_lookup [command]

Available Commands:
completion  Generate the autocompletion script for the specified shell
display     Human readable display of the registers and fields for the specified peripheral
dump        Dumps the SVD database
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
