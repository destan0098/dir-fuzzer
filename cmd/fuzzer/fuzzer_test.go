package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestReadLines tests reading lines from a file.
func TestReadLines(t *testing.T) {
	// Create a temporary file with sample content
	file, err := os.CreateTemp("", "wordlist.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Write some lines to the temporary file
	content := "test1\ntest2\ntest3\n"
	if _, err = file.WriteString(content); err != nil {
		t.Fatal(err)
	}

	// Read the lines from the file
	lines, err := readLines(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Check if the read lines match the expected content
	expected := []string{"test1", "test2", "test3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Expected line %d to be %s, got %s", i, expected[i], line)
		}
	}
}

// TestCheckURIs tests checking URIs and adding them to the seen URLs map.
func TestCheckURIs(t *testing.T) {
	// Initialize the seenurl map
	seenurl = make(map[string][][]string)

	// Create a slice to store response URIs
	var responseURIs []string

	// Create a mock HTTP server with predefined responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test1":
			w.WriteHeader(http.StatusOK)
		case "/test2":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Test with a base URL and wordlist
	baseURL := server.URL
	words := []string{"test1", "test2"}

	// Check URIs
	for _, word := range words {
		checkURIs(baseURL, word, &responseURIs)
	}

	// Validate the expected results
	if _, exists := seenurl[fmt.Sprintf("%s/test1", baseURL)]; !exists {
		t.Errorf("Expected /test1 to be in seen URLs")
	}

	if _, exists := seenurl[fmt.Sprintf("%s/test2", baseURL)]; exists {
		t.Errorf("Did not expect /test2 to be in seen URLs")
	}
}

// TestMainIntegration tests the main function's integration with other components.
func TestMainIntegration(t *testing.T) {
	// Set up initial data
	SiteUrl = "http://localhost/"
	Website = []string{SiteUrl}

	// Create a mock HTTP server to simulate requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test1":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Update the base URL to the mock server's URL
	Website[0] = server.URL

	// Set up a wordlist
	wordlistFile, err := os.CreateTemp("", "wordlist.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(wordlistFile.Name())

	wordlistContent := "test1\notherword\n"
	if _, err = wordlistFile.WriteString(wordlistContent); err != nil {
		t.Fatal(err)
	}

	// Reset the test seenurl map
	seenurl = make(map[string][][]string)

	// Call the main logic (not `main()`) to simulate the core functionality
	err = mainLogic(wordlistFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Validate the expected results
	if _, exists := seenurl[fmt.Sprintf("%s/test1", server.URL)]; !exists {
		t.Errorf("Expected /test1 to be in seen URLs")
	}

	if _, exists := seenurl[fmt.Sprintf("%s/otherword", server.URL)]; exists {
		t.Errorf("Did not expect /otherword to be in seen URLs")
	}
}

// Refactor the core logic into a testable function
func mainLogic(wordlistPath string) error {
	// Read the wordlist from a file
	wordlist, err := readLines(wordlistPath)
	if err != nil {
		return err
	}

	// Create a slice to store the response URIs with status code 200
	for i := 0; i < 3; i++ {
		for _, url := range Website {
			for _, word := range wordlist {
				word = strings.TrimRight(word, "\r\n")

				checkURIs(url, word, &Website)
			}
		}
	}

	// Print the response URIs
	fmt.Println("Response URIs with status code 200:")
	for uri, method := range seenurl {
		fmt.Printf("URI: %s, Method: %s and Status Code: %s\n", uri, method[0], method[1])
	}

	return nil
}
