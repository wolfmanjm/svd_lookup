/*
Copyright © 2025 Jim Morris <morris@wolfman.com>
*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"
	svd_lookup "github.com/wolfmanjm/svd_lookup/internal"
)

var verbose bool
var cwd string
var database string
var periph string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "svd_lookup",
	Short: "queries a SVD database",
	Long: `Query a SVD database in various ways.
	Depending on the subcommand it can generate various code sequences to access the peripherals and registers
	or display the available peripherals and/or registers in a human readable way.
	`,
	PersistentPreRunE: pre_run,
	PersistentPostRun: post_run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pre_run(cmd *cobra.Command, args []string) error {
	if cmd.Name() != "convert" {
		if cwd != "" {
			svd_lookup.SetSearchPath(cwd)
		}
		if database != "" {
			svd_lookup.SetDatabase(database)
		}
		if verbose {
			svd_lookup.SetVerbose()
		}
		return svd_lookup.OpenDatabase()

	} else {
		return nil
	}
}

func post_run(cmd *cobra.Command, args []string) {
	if cmd.Name() != "convert" {
		svd_lookup.CloseDatabase()
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&cwd, "curdir", "c", "", "set the current directory for db search")
	rootCmd.PersistentFlags().StringVarP(&database, "database", "d", "", "use the named database")
	rootCmd.MarkFlagsMutuallyExclusive("curdir", "database")
}


