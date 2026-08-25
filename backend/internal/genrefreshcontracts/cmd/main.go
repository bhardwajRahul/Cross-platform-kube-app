package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/luxury-yacht/app/backend/internal/genrefreshcontracts"
)

func main() {
	out := flag.String("out", "", "generated TypeScript output path")
	goPolicyOut := flag.String("go-policy-out", "", "generated Go domain-policy output path")
	flag.Parse()
	if *out == "" || *goPolicyOut == "" {
		fmt.Fprintln(os.Stderr, "-out and -go-policy-out are required")
		os.Exit(2)
	}

	generated, err := genrefreshcontracts.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "render refresh contracts: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write refresh contracts: %v\n", err)
		os.Exit(1)
	}

	goPolicy, err := genrefreshcontracts.RenderGoPolicy()
	if err != nil {
		fmt.Fprintf(os.Stderr, "render refresh domain policy: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*goPolicyOut, goPolicy, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write refresh domain policy: %v\n", err)
		os.Exit(1)
	}
}
