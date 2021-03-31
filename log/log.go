package log

import "log"

func Errorf(f string, v ...interface{}) {
	log.Printf(f, v...)
}

func Notice(f string, v ...interface{}) {
	log.Printf(f, v...)
}
