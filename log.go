package main

import (
	"io"
	"log"
	"os"
)

var logger *log.Logger

func openLogFile(filepath string) error {

	var w io.Writer

	if filepath == "stdout" {
		w = os.Stdout
	} else {

		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		w = f
	}

	logger = log.New(w, "", log.LstdFlags)

	return nil
}

func CMDLogln(user, host string, v ...interface{}) {

	logger.Println(append([]interface{}{user, ">", host, ">"}, v...)...)
}
