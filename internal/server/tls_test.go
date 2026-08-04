package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTLSMaterialGenerates(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "cert.pem")
	key := filepath.Join(dir, "key.pem")

	c, k, gen, err := EnsureTLSMaterial(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	if !gen {
		t.Fatal("expected generation")
	}
	if c != cert || k != key {
		t.Fatalf("paths %s %s", c, k)
	}
	if _, err := os.Stat(cert); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Fatal(err)
	}

	// Second call reuses files.
	_, _, gen2, err := EnsureTLSMaterial(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	if gen2 {
		t.Fatal("should not regenerate")
	}
}
