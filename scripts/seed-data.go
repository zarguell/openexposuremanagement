package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultAPIURL = "http://localhost:8080"

type Config struct {
	APIURL   string
	DataDir  string
	DemoMode bool
	Verbose  bool
}

func main() {
	config := Config{}

	flag.StringVar(&config.APIURL, "api-url", defaultAPIURL, "API base URL")
	flag.StringVar(&config.DataDir, "data-dir", "sample-data", "Directory containing sample data files")
	flag.BoolVar(&config.DemoMode, "demo", false, "Run in demo mode (no authentication)")
	flag.BoolVar(&config.Verbose, "verbose", true, "Enable verbose logging")
	flag.Parse()

	// Always enable verbose logging for demo mode
	if config.DemoMode || config.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	fmt.Printf("🌱 Starting OEM data seeding\n")
	fmt.Printf("📡 API URL: %s\n", config.APIURL)
	fmt.Printf("📁 Data directory: %s\n", config.DataDir)
	if config.DemoMode {
		fmt.Printf("🔓 Demo mode: Authentication disabled\n")
	}
	fmt.Println()

	// Check if data directory exists
	if _, err := os.Stat(config.DataDir); os.IsNotExist(err) {
		log.Fatalf("❌ Data directory does not exist: %s", config.DataDir)
	}

	// Find all JSON files in the data directory
	files, err := filepath.Glob(filepath.Join(config.DataDir, "*.json"))
	if err != nil {
		log.Fatalf("❌ Failed to find data files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("❌ No JSON files found in %s", config.DataDir)
	}

	fmt.Printf("📋 Found %d sample data files\n", len(files))

	// Process each file
	totalProcessed := 0
	totalAssets := 0
	totalFindings := 0

	for _, file := range files {
		fmt.Printf("\n📄 Processing %s...\n", filepath.Base(file))

		assets, findings, err := processFile(file, config)
		if err != nil {
			log.Printf("❌ Failed to process %s: %v", file, err)
			continue
		}

		totalProcessed++
		totalAssets += assets
		totalFindings += findings

		fmt.Printf("✅ Processed %d assets, %d findings\n", assets, findings)
	}

	fmt.Printf("\n🎉 Seeding complete!\n")
	fmt.Printf("📊 Summary: %d files processed, %d assets, %d findings\n", totalProcessed, totalAssets, totalFindings)
	fmt.Printf("🔄 Data is idempotent - run again to verify no duplicates created\n")
}

func processFile(filename string, config Config) (int, int, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON to count assets/findings
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	findings, ok := payload["findings"].([]interface{})
	if !ok {
		return 0, 0, fmt.Errorf("no findings array in payload")
	}

	// Count unique assets
	assetMap := make(map[string]bool)
	for _, finding := range findings {
		findingMap, ok := finding.(map[string]interface{})
		if !ok {
			continue
		}

		asset, ok := findingMap["asset"].(map[string]interface{})
		if !ok {
			continue
		}

		hostname, _ := asset["hostname"].(string)
		if hostname != "" {
			assetMap[hostname] = true
		}
	}

	assetCount := len(assetMap)
	findingCount := len(findings)

	// Send to API
	url := fmt.Sprintf("%s/api/v1/ingest/vm/findings", config.APIURL)

	if config.Verbose {
		log.Printf("POST %s", url)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add demo mode header if needed
	if config.DemoMode {
		req.Header.Set("X-Demo-Mode", "true")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	if config.Verbose {
		log.Printf("Response status: %s", resp.Status)
		log.Printf("Response body: %s", string(body))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to verify success
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if status, ok := response["status"].(string); !ok || status != "success" {
		return 0, 0, fmt.Errorf("API returned non-success status: %s", string(body))
	}

	return assetCount, findingCount, nil
}
