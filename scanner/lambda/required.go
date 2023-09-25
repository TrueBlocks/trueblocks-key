package main

import (
	"errors"
	"os"
)

var BUCKET_NAME string

func init() {
	var ok bool
	if BUCKET_NAME, ok = os.LookupEnv("SCNR_BUCKET_NAME"); !ok {
		panic(errors.New("env variable SCNR_BUCKET_NAME required"))
	}
}
