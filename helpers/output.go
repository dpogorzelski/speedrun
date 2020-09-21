package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type progress struct {
	spinner *spinner.Spinner
	command func(a ...interface{}) string
}

func NewProgress() *progress {
	o := &progress{}
	o.spinner = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	o.command = color.New(color.FgCyan).SprintFunc()
	return o
}

func (o *progress) Start(msg string) {
	t := color.GreenString("✓")
	o.spinner.Suffix = " " + msg
	o.spinner.FinalMSG = fmt.Sprintf("%s %s\n", t, msg)
	o.spinner.Start()
}

func (o *progress) Stop() {
	o.spinner.Stop()
}

func (o *progress) Error(err error) {
	t := color.RedString("-")
	o.spinner.FinalMSG = fmt.Sprintf("%s%s: %s\n", t, o.spinner.Suffix, err)
	o.spinner.Stop()
}

func (o *progress) Failure(msg string) {
	t := color.YellowString("X")
	o.spinner.FinalMSG = fmt.Sprintf("%s%s: %s\n", t, o.spinner.Suffix, msg)
	o.spinner.Stop()
	os.Exit(0)
}

func Error(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
