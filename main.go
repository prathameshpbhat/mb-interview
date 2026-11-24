package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	mu   sync.Mutex
	data map[string]string
}

func (c Cache) Get(path string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	h, ok := c.data[path]
	return h, ok
}

func (c *Cache) Set(path, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[path] = hash
}

var cache = &Cache{data: map[string]string{}}

func hashFile(path string) (string, error) {
	if h, ok := cache.Get(path); ok {
		return h, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	h := hex.EncodeToString(sum[:])
	cache.Set(path, h)
	return h, nil
}

type result struct {
	path string
	hash string
	err  error
}

// Dedupe returns groups of files in dir that have identical content.
func Dedupe(dir string) (map[string][]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}

	results := make(chan result)
	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			h, err := hashFile(p)
			results <- result{p, h, err}
		}(f)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	groups := map[string][]string{}
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		groups[r.hash] = append(groups[r.hash], r.path)
	}

	for h, g := range groups {
		if len(g) < 2 {
			delete(groups, h)
		}
	}
	return groups, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: dedupe <dir>")
		os.Exit(1)
	}
	groups, err := Dedupe(os.Args[1])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	for h, g := range groups {
		fmt.Println(h[:12], g)
	}
}
