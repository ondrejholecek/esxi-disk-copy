package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func Infof(v ...interface{}) {
	msg := v[0].(string)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	msg = time.Now().Format("2006/01/02 15:04:05 ") + msg
	fmt.Fprintf(os.Stdout, msg, v[1:]...)
}

func Fatalf(v ...interface{}) {
	msg := v[0].(string)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	msg = time.Now().Format("2006/01/02 15:04:05 ") + msg
	fmt.Fprintf(os.Stderr, msg, v[1:]...)
	os.Exit(1)
}
