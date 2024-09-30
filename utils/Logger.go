package utils

import (
	"log"
	"os"
)

type LoggerWrapper struct {
	log.Logger
}

var Logger LoggerWrapper

func init() {
	Logger.Logger = *log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}
