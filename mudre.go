package main

import (
	"log"

	"github.com/RETr4ce/mudre/cmd"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func main() {
	cmd.Execute()
}
