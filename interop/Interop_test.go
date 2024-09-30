package Interop

import (
	"os"
	"testing"
)

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestExtractLinksFromURL(t *testing.T) {
	// Load Python
	err := LoadPython()
	if err != nil {
		t.Fatalf("LoadPython failed with error: %v", err)
	}
	// Extract links from www.microsoft.com
	url := "https://www.microsoft.com"
	// call the function with "depth" 1.
	links, err := ExtractLinksFromURL(url, 1)
	if err != nil {
		t.Fatalf("ExtractLinksFromURL failed with error: %v", err)
	}
	// make sure you got some links
	if len(links) == 0 {
		t.Fatalf("ExtractLinksFromURL returned no links")
	}
	// print the links
	// Notice, Log and Logf are printed only when using "-v" switch
	t.Logf("links: %v\n", links)
}

func TestChordDHT(t *testing.T) {
	// Load JVM
	err := LoadJVM()
	if err != nil {
		t.Fatalf("LoadJVM failed with error: %v", err)
	}

	// Create a new ChordDHT
	root, err := NewChordDHT("root", 9009)
	if err != nil {
		t.Fatalf("NewChordDHT failed with error: %v", err)
	}

	// Join the ChordDHT two more nodes
	node2, err := JoinChordDHT("node2", "root", 9009)
	if err != nil {
		t.Fatalf("JoinChordDHT node2 failed with error: %v", err)
	}

	node3, err := JoinChordDHT("node3", "root", 9009)
	if err != nil {
		t.Fatalf("JoinChordDHT node3 failed with error: %v", err)
	}

	// Put three key-value pairs in the ChordDHT using
	// different nodes
	err = root.Set("key1", "value1")
	if err != nil {
		t.Fatalf("root.Set failed with error: %v", err)
	}

	err = node2.Set("key2", "value2")
	if err != nil {
		t.Fatalf("node2.Set failed with error: %v", err)
	}

	err = node3.Set("key3", "value3")
	if err != nil {
		t.Fatalf("node3.Set failed with error: %v", err)
	}

	// Get all keys in the ChordDHT and make sure they are as expected
	keys, err := node2.GetAllKeys()
	if err != nil {
		t.Fatalf("node2.GetAllKeys failed with error: %v", err)
	}

	if len(keys) != 3 {
		t.Fatalf("node2.GetAllKeys returned %d keys, expected 3", len(keys))
	}

	// make sure "key1", "key2", and "key3" are in the keys
	for _, key := range keys {
		if key == "key1" || key == "key2" || key == "key3" {
			continue
		}
		t.Fatalf("node2.GetAllKeys returned unexpected key: %s", key)
	}

	// Get the value for one of the keys
	value, err := node3.Get("key2")
	if err != nil {
		t.Fatalf("node3.Get failed with error: %v", err)
	}

	if value != "value2" {
		t.Fatalf("node3.Get returned unexpected value: %s", value)
	}

	// Remove the key-value pair for one of the keys
	err = root.Delete("key2")
	if err != nil {
		t.Fatalf("root.Delete failed with error: %v", err)
	}

	// Get the value for the removed key
	value, err = node3.Get("key2")
	if err != nil {
		t.Fatalf("node3.Get failed with error: %v", err)
	}

	if value != "" {
		t.Fatalf("node3.Get returned value for a deleted key: %s", value)
	}

	// Get all keys in the ChordDHT make sure the keys are as expected
	keys, err = root.GetAllKeys()
	if err != nil {
		t.Fatalf("root.GetAllKeys failed with error: %v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("root.GetAllKeys returned %d keys, expected 2", len(keys))
	}

	// make sure "key1" and "key3" are in the keys
	for _, key := range keys {
		if key == "key1" || key == "key3" {
			continue
		}
		t.Fatalf("root.GetAllKeys returned unexpected key: %s", key)
	}
}
