package utils

import (
	"log"
)

// LogPlainln prints a line without a prefix using the log package
func LogPlainln(v ...any) {
	flags := log.Flags()
	log.SetFlags(0)
	log.Println(v...)
	log.SetFlags(flags)
}
