package logx

import "log"

// Plainln prints a line without a prefix using the log package
func Plainln(v ...any) {
	flags := log.Flags()
	log.SetFlags(0)
	log.Println(v...)
	log.SetFlags(flags)
}
