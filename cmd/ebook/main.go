package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anwar/ai-content-engine/internal/affiliate"
	"github.com/anwar/ai-content-engine/internal/env"
)

func main() {
	env.Load()
	mdPath := os.Getenv("BLOG_PATH")
	pdfPath := os.Getenv("PDF_PATH")
	title := os.Getenv("EBOOK_TITLE")

	if mdPath == "" || pdfPath == "" {
		log.Fatal("BLOG_PATH and PDF_PATH required")
	}
	if title == "" {
		title = "Daily AI Insights"
	}

	// Step 1: Inject affiliate links into blog post
	blog, err := os.ReadFile(mdPath)
	if err != nil {
		log.Fatalf("read blog: %v", err)
	}

	affiliateContent, err := affiliate.Inject(string(blog), nil)
	if err != nil {
		log.Printf("affiliate inject warning: %v", err)
		affiliateContent = string(blog)
	}

	affiliatePath := filepath.Join(filepath.Dir(mdPath), "blog_with_affiliates.md")
	os.WriteFile(affiliatePath, []byte(affiliateContent), 0644)
	fmt.Println("  ✓ Affiliate links injected")

	// Step 2: Convert to PDF using pandoc (pre-installed on ubuntu-latest)
	os.MkdirAll(filepath.Dir(pdfPath), 0755)

	cmd := exec.Command("pandoc", affiliatePath, "-o", pdfPath, "--pdf-engine=xelatex")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("  ⚠ pandoc failed (texlive may be missing), trying markdown2pdf: %v\n", err)
		cmd2 := exec.Command("markdown2pdf", affiliatePath, "-o", pdfPath)
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err2 := cmd2.Run(); err2 != nil {
			log.Fatalf("both pandoc and markdown2pdf failed: %v", err2)
		}
	}

	fmt.Printf("  ✓ PDF created: %s\n", pdfPath)

	// Step 3: Create Gumroad product (using gumroad-cli Go tool)
	if token := os.Getenv("GUMROAD_ACCESS_TOKEN"); token != "" {
		gumCmd := exec.Command("gumroad", "products", "create",
			"--name", title,
			"--price", "4.99",
			"--type", "ebook",
			"--file", pdfPath,
			"--file-name", title+".pdf",
			"--json", "--no-input",
		)
		gumCmd.Env = append(os.Environ(), "GUMROAD_ACCESS_TOKEN="+token)
		output, err := gumCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  ⚠ Gumroad create (may already exist): %v\n", err)
		}
		fmt.Printf("  ✓ Gumroad product created\n  %s\n", string(output))
	}
}
