package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Progress represents a progress spinner for a given command
type Progress struct {
	spinner *spinner.Spinner
	command func(a ...interface{}) string
}

// NewProgress creates a new progress indicator/spinner
func NewProgress() *Progress {
	o := &Progress{}
	o.spinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	o.command = color.New(color.FgCyan).SprintFunc()
	return o
}

// Start sets the default "success" condition tag that will be displayed when the spinner is stopped
func (o *Progress) Start(msg string) {
	tag := green("✓")
	o.spinner.Suffix = " " + msg
	o.spinner.FinalMSG = fmt.Sprintf("%s %s\n", tag, msg)
	o.spinner.Start()
}

// Stop stops the spinner
func (o *Progress) Stop() {
	o.spinner.Stop()
}

// Failure changes the spinner tag to the one associated with "failure" conditions and exits
func (o *Progress) Failure(msg string) {
	tag := yellow("X")
	o.spinner.FinalMSG = fmt.Sprintf("%s%s: %s\n", tag, o.spinner.Suffix, msg)
	o.spinner.Stop()
	os.Exit(0)
}

func (o *Progress) Error(err error) {
	tag := red("-")
	o.spinner.FinalMSG = fmt.Sprintf("%s%s: %s\n", tag, o.spinner.Suffix, err)
	o.spinner.Stop()
}
