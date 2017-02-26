package cmd

import "github.com/fatih/color"

var (
	warning = color.New(color.FgRed)
	notice  = color.New(color.FgGreen)
	debug   = color.New(color.FgBlue)
)

type command struct {
	Verbose bool   `long:"verbose" description:"show verbose debug information"`
	Root    string `long:"root" default:"." description:"root of the driver"`
}
