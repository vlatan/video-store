package color

import "fmt"

// Define ANSI color codes
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

func Yellow(msg string) {
	fmt.Println(yellow + msg + reset)
}

func Red(msg string) {
	fmt.Println(red + msg + reset)
}
