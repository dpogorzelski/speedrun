package utils

import (
	"fmt"
	"os"

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

// Error is a shortcut function that prints the error message and exits
func Error(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
