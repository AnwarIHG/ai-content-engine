package env

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Load reads .env from the project root.
// Walks up from cwd looking for .env. Falls back silently if not found.
func Load() {
	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 5; i++ {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("  ⚠ Error loading %s: %v", path, err)
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// .env not found — that's fine, env vars may be set another way
}
