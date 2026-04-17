package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dobby/filemanager/internal/domain/license"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randSegment() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, 4)
	for i, v := range b {
		out[i] = charset[int(v)%len(charset)]
	}
	return string(out), nil
}

func generateKeys(n int) ([]string, error) {
	if n < 0 {
		return nil, errors.New("count must be >= 0")
	}
	validator := license.NewLicenseKeyValidator()
	seen := make(map[string]struct{}, n)
	keys := make([]string, 0, n)
	for len(keys) < n {
		p1, err := randSegment()
		if err != nil {
			return nil, err
		}
		p2, err := randSegment()
		if err != nil {
			return nil, err
		}
		key := validator.GenerateKey(p1, p2)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys, nil
}

func main() {
	n := flag.Int("n", 10000, "number of keys to generate")
	o := flag.String("o", "", "output file path (default: stdout)")
	flag.Parse()

	if *n < 0 {
		fmt.Fprintln(os.Stderr, "error: -n must be >= 0")
		os.Exit(1)
	}

	keys, err := generateKeys(*n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var w *bufio.Writer
	if *o == "" {
		w = bufio.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(*o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = bufio.NewWriter(f)
	}

	for _, k := range keys {
		fmt.Fprintln(w, k)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
