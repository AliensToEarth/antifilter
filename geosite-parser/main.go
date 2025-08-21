package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

var (
	configFile = flag.String("c", "../geosite-antifilter.json", "Path to the config file")
	help       = flag.Bool("h", false, "Show help message")
)

type Config struct {
	Input  []InputConfig  `json:"input"`
	Output []OutputConfig `json:"output"`
}

type InputConfig struct {
	Type   string                 `json:"type"`
	Action string                 `json:"action"`
	Args   map[string]interface{} `json:"args"`
}

type OutputConfig struct {
	Type   string                 `json:"type"`
	Action string                 `json:"action"`
	Args   map[string]interface{} `json:"args"`
}

func main() {
	flag.Parse()

	if *help {
		fmt.Println("GeoSite Parser Utility")
		fmt.Println("Usage: go run main.go [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -c string")
		fmt.Println("        Path to the config file (default \"../geosite-antifilter.json\")")
		fmt.Println("  -h    Show help message")
		fmt.Println()
		fmt.Println("This utility processes geosite data using the specified configuration file.")
		fmt.Println("It runs after 'make geosite' to create filtered geosite output files.")
		return
	}

	// Check if config file exists
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", *configFile)
	}

	// Read and parse config file
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	fmt.Printf("Using config file: %s\n", *configFile)

	// Process input
	var geoSiteList *router.GeoSiteList
	for _, input := range config.Input {
		if input.Type == "readGeoSiteDat" && input.Action == "add" {
			path, ok := input.Args["path"].(string)
			if !ok {
				log.Fatalf("Invalid path in input config")
			}

			// Make path relative to current working directory if not absolute
			if !filepath.IsAbs(path) {
				// The path in config is relative to the project root, not the config file
				path = filepath.Clean(path)
			}

			fmt.Printf("Reading geosite data from: %s\n", path)
			geoSiteList, err = readGeoSiteDat(path)
			if err != nil {
				log.Fatalf("Failed to read geosite data: %v", err)
			}
			fmt.Printf("Loaded %d geosite entries\n", len(geoSiteList.Entry))
		}
	}

	if geoSiteList == nil {
		log.Fatalf("No geosite data loaded")
	}

	// Process output
	for _, output := range config.Output {
		if output.Type == "v2rayGeoSiteDat" && output.Action == "output" {
			err := processOutput(geoSiteList, output)
			if err != nil {
				log.Fatalf("Failed to process output: %v", err)
			}
		}
	}

	fmt.Println("GeoSite processing completed successfully!")
}

func readGeoSiteDat(path string) (*router.GeoSiteList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var geoSiteList router.GeoSiteList
	if err := proto.Unmarshal(data, &geoSiteList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %v", err)
	}

	return &geoSiteList, nil
}

func processOutput(geoSiteList *router.GeoSiteList, output OutputConfig) error {
	outputDir, ok := output.Args["outputDir"].(string)
	if !ok {
		return fmt.Errorf("invalid outputDir in output config")
	}

	outputName, ok := output.Args["outputName"].(string)
	if !ok {
		return fmt.Errorf("invalid outputName in output config")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Filter wanted lists if specified
	var filteredList *router.GeoSiteList
	if wantedListInterface, exists := output.Args["wantedList"]; exists {
		wantedListSlice, ok := wantedListInterface.([]interface{})
		if !ok {
			return fmt.Errorf("invalid wantedList format")
		}

		wantedList := make([]string, len(wantedListSlice))
		for i, v := range wantedListSlice {
			wantedList[i], ok = v.(string)
			if !ok {
				return fmt.Errorf("invalid wantedList item")
			}
		}

		filteredList = filterGeoSiteList(geoSiteList, wantedList)
		fmt.Printf("Filtered to %d entries from wanted list: %v\n", len(filteredList.Entry), wantedList)
	} else {
		filteredList = geoSiteList
	}

	// Sort the list for reproducible output
	sort.SliceStable(filteredList.Entry, func(i, j int) bool {
		return filteredList.Entry[i].CountryCode < filteredList.Entry[j].CountryCode
	})

	// Marshal to protobuf
	data, err := proto.Marshal(filteredList)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %v", err)
	}

	// Write output file
	outputPath := filepath.Join(outputDir, outputName)
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func filterGeoSiteList(geoSiteList *router.GeoSiteList, wantedList []string) *router.GeoSiteList {
	wantedMap := make(map[string]bool)
	for _, name := range wantedList {
		// Convert to uppercase to match the actual country codes in geosite data
		wantedMap[strings.ToUpper(name)] = true
	}

	filtered := &router.GeoSiteList{}
	for _, entry := range geoSiteList.Entry {
		if wantedMap[entry.CountryCode] {
			filtered.Entry = append(filtered.Entry, entry)
		}
	}

	return filtered
}