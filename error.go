package main

import (
	"fmt"
	"html"
	"os"
	"runtime/debug"
	"time"
)

func fatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		msg := fmt.Sprintf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Printf("%s", msg)
		fmt.Fprintf(os.Stderr, "%s", msg)
		_ = sendStatusEmail("<b><p style=\"color:red\">ES monitor error:</p></b>\n" + html.EscapeString(msg))
		panic("stacktrace")
	}
}

func fatalf(f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...))
}
