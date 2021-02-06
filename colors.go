package main

import (
	"github.com/fatih/color"
)

func green(s string) string {
	return color.GreenString(s)
}

func yellow(s string) string {
	return color.YellowString(s)
}

func red(s string) string {
	return color.RedString(s)
}
