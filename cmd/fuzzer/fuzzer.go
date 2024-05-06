package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var SiteUrl = ""

var Website []string

var seenurl map[string][][]string // Declare the map to track unique URIs
func withPipe() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		SiteUrl = scanner.Text()
		SiteUrl = strings.TrimPrefix(SiteUrl, "\ufeff")
		if len(Website) == 0 {

			Website = append(Website, SiteUrl)
		}
	}

}
func writeJSONFile(file *os.File, seenurl map[string][][]string) {
	jsonData, err := json.MarshalIndent(seenurl, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := file.Write(jsonData); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSon File Created \n")
}

// detect file format to save output file
func writeToFile(seenurl map[string][][]string, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	writeJSONFile(file, seenurl)

}

func main() {
	withPipe()
	seenurl = make(map[string][][]string) // Proper initialization of seenurl

	// Read the wordlist from a file
	wordlist, err := readLines("wordlist.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Create a slice to store the response URIs with status code 200

	for i := 0; i < 2; i++ {
		for _, url := range Website {
			for _, word := range wordlist {
				word = strings.TrimRight(word, "\r\n")

				checkURIs(url, word, &Website)
			}
		}
	}

	// Print the response URIs
	fmt.Println("Finded URIs")
	writeToFile(seenurl, "output.json")
	//print seenurl all value
	for k, v := range seenurl {
		fmt.Printf("URI: %s, Methods: %v\n", k, v)
	}

}
func TestMethodds(url, method string, responseURIs *[]string) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		seenurl[url] = append(seenurl[url], []string{method, resp.Status}) // Use the constructed URI as key
		*responseURIs = append(*responseURIs, url)
		//send url to other application stdin
		fmt.Fprintln(os.Stdout, url)
		fmt.Printf("%s %s - Status: %s\n", method, url, resp.Status)
		fmt.Println(strings.Repeat("-", 20))

	}

}

func checkURIs(baseURL, word string, responseURIs *[]string) {

	baseURL = strings.TrimSuffix(baseURL, "/")

	makeURI := fmt.Sprintf("%s/%s", baseURL, word)

	// Prevent duplicate checking
	if _, exists := seenurl[makeURI]; exists {
		return
	}

	TestMethodds(makeURI, http.MethodGet, responseURIs)
	TestMethodds(makeURI, http.MethodPost, responseURIs)
	TestMethodds(makeURI, http.MethodDelete, responseURIs)
	TestMethodds(makeURI, http.MethodPut, responseURIs)
	TestMethodds(makeURI, http.MethodOptions, responseURIs)

}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
