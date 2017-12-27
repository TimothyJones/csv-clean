package main

import (
	"flag"
	"os"
)

type CleanConfig struct {
	correctSlashedQuotes bool
}

func main() {
	correctSlashedQuotes := flag.Bool("fixSlashedQuotes", true, "Whether to correct '\"' quotes into '\"\"'")

	flag.Parse()

	config := &CleanConfig{
		correctSlashedQuotes: *correctSlashedQuotes,
	}

	Clean(os.Stdin, os.Stdout, config)
}
