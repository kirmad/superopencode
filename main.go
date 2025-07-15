package main

import (
	"github.com/kirmad/superopencode/cmd"
	"github.com/kirmad/superopencode/internal/logging"
)

func main() {
	defer logging.RecoverPanic("main", func() {
		logging.ErrorPersist("Application terminated due to unhandled panic")
	})

	cmd.Execute()
}
