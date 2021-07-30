package colors

import (
	"github.com/jwalton/gchalk"
)

// Green returns a green string
func Green(s string) string {
	return gchalk.Green(s)
}

// Yellow returns a green string
func Yellow(s string) string {
	return gchalk.Yellow(s)
}

// Red returns a green string
func Red(s string) string {
	return gchalk.Red(s)
}

// Blue returns a blue string
func Blue(s string) string {
	return gchalk.Blue(s)
}

// White returns a white string
func White(s string) string {
	return gchalk.White(s)
}

// Purple returns a purple string
func Purple(s string) string {
	return gchalk.RGB(196, 16, 224)(s)
}
