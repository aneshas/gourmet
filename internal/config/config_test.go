package config

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestFoo(t *testing.T) {
	f, err := os.Open(filepath.Join("test_fixture", "example.toml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := parse(f)

	log.Println(err)
	log.Println(cfg)
}

// TODO - Add integrity test eg. loop through locations and check for corresponding upstreams exist
