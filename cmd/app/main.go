// Package main is the entry point for the application.
// It uses Cobra for CLI command handling and Wire for dependency injection.
package main

import (
	"go-clean-arch/cmd/app/cmd"
)

func main() {
	cmd.Execute()
}
