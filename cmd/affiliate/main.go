package main

import (
	"fmt"
	"log"
	"os"

	"github.com/anwar/ai-content-engine/internal/affiliate"
	"github.com/anwar/ai-content-engine/internal/env"
)

func main() {
	env.Load()
	input := os.Getenv("AFFILIATE_INPUT")
	output := os.Getenv("AFFILIATE_OUTPUT")

	if input == "" || output == "" {
		log.Fatal("AFFILIATE_INPUT and AFFILIATE_OUTPUT required")
	}

	data, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("read input: %v", err)
	}

	result, err := affiliate.Inject(string(data), nil)
	if err != nil {
		log.Fatalf("affiliate inject: %v", err)
	}

	if err := os.WriteFile(output, []byte(result), 0644); err != nil {
		log.Fatalf("write output: %v", err)
	}

	fmt.Printf("  ✓ Affiliate links injected: %s\n", output)
}
