package main

import (
	"log"
	"path/filepath"
	"runtime"
)

func basePath() string {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		log.Panic("Caller error")
	}
	return filepath.Dir(b)
}
func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}