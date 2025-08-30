package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings" // Added for strings.HasSuffix
	"sync"    // Added for mutex
	"time"    // Added for time.Tick

	"github.com/fingon/ddo-trove-ui/db"
	"github.com/fingon/ddo-trove-ui/templates"
)

// Helper functions for pagination
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getPaginationRange(currentPage, totalPages int) (int, int) {
	maxButtonsToShow := 10
	startPage := currentPage - 4
	endPage := currentPage + 5

	// Adjust if the start page goes below 1
	if startPage < 1 {
		startPage = 1
		if totalPages < maxButtonsToShow {
			endPage = totalPages
		} else {
			endPage = maxButtonsToShow
		}
	}

	// Adjust if the end page goes above totalPages
	if endPage > totalPages {
		endPage = totalPages
		if totalPages > maxButtonsToShow {
			startPage = totalPages - maxButtonsToShow + 1
		} else {
			startPage = 1
		}
	}

	// Final check to ensure startPage is not less than 1 after adjustments
	if startPage < 1 {
		startPage = 1
	}

	return startPage, endPage
}

const (
	itemsPerPage = 100
	defaultPage  = 1
)

var (
	inputDir string
	// Protected by itemsMutex
	allItems     *db.AllItems
	fileModTimes map[string]time.Time
	itemsMutex   sync.RWMutex
)

func init() {
	flag.StringVar(&inputDir, "input", "", "Directory containing JSON files (e.g., example/local)")
}

// monitorAndReloadItems checks the input directory for changes and reloads items if necessary.
func monitorAndReloadItems(dirPath string) {
	log.Println("Checking for file changes...")
	currentFileModTimes := make(map[string]time.Time)

	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory for monitoring: %v", err)
		return
	}

	needsReload := false

	// Acquire read lock to compare with current global fileModTimes
	itemsMutex.RLock()
	defer itemsMutex.RUnlock() // Release read lock after comparison

	// Check for new files or modified files
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Warning: Failed to get file info for %s during monitoring: %v\n", filePath, err)
			continue
		}
		currentFileModTimes[filePath] = info.ModTime()

		oldModTime, exists := fileModTimes[filePath]
		if !exists || info.ModTime().After(oldModTime) {
			log.Printf("Detected change in file: %s (old: %v, new: %v)", filePath, oldModTime, info.ModTime())
			needsReload = true
			break
		}
	}

	// If no changes detected yet, check for deleted files or file count changes
	if !needsReload {
		if len(currentFileModTimes) != len(fileModTimes) {
			log.Println("Detected file count change (possibly deleted/added files). Reloading.")
			needsReload = true
		} else {
			// If file counts are the same, but a file was renamed, this won't catch it
			// unless we iterate through old fileModTimes and check if they exist in currentFileModTimes
			for oldPath := range fileModTimes {
				if _, exists := currentFileModTimes[oldPath]; !exists {
					log.Printf("Detected deleted file: %s. Reloading.", oldPath)
					needsReload = true
					break
				}
			}
		}
	}

	if needsReload {
		log.Println("Reloading all items due to detected changes...")
		// Acquire write lock to update global allItems and fileModTimes
		itemsMutex.RUnlock() // Release read lock before acquiring write lock
		itemsMutex.Lock()
		defer itemsMutex.Unlock() // Release write lock

		newAllItems, newFileModTimes, err := db.LoadItemsFromDir(dirPath)
		if err != nil {
			log.Printf("Error reloading items: %v", err)
			// Re-acquire read lock if we failed to reload, to maintain consistent state for defer RUnlock
			itemsMutex.RLock()
			return
		}
		allItems = newAllItems
		fileModTimes = newFileModTimes
		log.Printf("Reload complete. Loaded %d items.", len(allItems.Items))
	} else {
		log.Println("No file changes detected.")
	}
}

func main() {
	flag.Parse()

	if inputDir == "" {
		log.Fatal("Error: --input directory is required.")
	}

	// Ensure the input directory exists and is valid
	absPath, err := filepath.Abs(inputDir)
	if err != nil {
		log.Fatalf("Error resolving absolute path for input directory: %v", err)
	}
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		log.Fatalf("Error: Input directory '%s' does not exist.", absPath)
	}
	if !info.IsDir() {
		log.Fatalf("Error: Input path '%s' is not a directory.", absPath)
	}

	// Initial load of items and file modification times
	itemsMutex.Lock()
	allItems, fileModTimes, err = db.LoadItemsFromDir(absPath)
	if err != nil {
		log.Fatalf("Error loading items: %v", err)
	}
	log.Printf("Initial load: Loaded %d items.", len(allItems.Items))
	itemsMutex.Unlock()

	// Start goroutine to monitor files and reload every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			monitorAndReloadItems(absPath)
		}
	}()

	// Serve static files (CSS, images, etc.)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// HTTP Handlers
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Acquire read lock before accessing allItems
		itemsMutex.RLock()
		defer itemsMutex.RUnlock()

		itemType := r.URL.Query().Get("itemType")
		if itemType == "" {
			itemType = "All" // Default to "All" if not specified
		}

		itemSubType := r.URL.Query().Get("itemSubType")
		if itemSubType == "" {
			itemSubType = "All" // Default to "All" if not specified
		}

		characterName := r.URL.Query().Get("characterName")
		if characterName == "" {
			characterName = "All" // Default to "All" if not specified
		}

		nameSearch := r.URL.Query().Get("nameSearch")

		minLevelStr := r.URL.Query().Get("minLevel")
		minLevel, err := strconv.Atoi(minLevelStr)
		if err != nil || minLevel < 0 {
			minLevel = 0 // Default minimum level
		}

		maxLevelStr := r.URL.Query().Get("maxLevel")
		maxLevel, err := strconv.Atoi(maxLevelStr)
		if err != nil || maxLevel < 0 {
			maxLevel = 40 // Default maximum level
		}

		// New: EquipsTo filter parameter
		equipsTo := r.URL.Query().Get("equipsTo")
		if equipsTo == "" {
			equipsTo = "All" // Default to "All" if not specified
		}

		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = defaultPage // Default to page 1
		}

		filteredItems := db.FilterItems(allItems.Items, itemType, itemSubType, characterName, nameSearch, minLevel, maxLevel, equipsTo) // Pass equipsTo
		log.Printf("Initial load/filter by itemType='%s', itemSubType='%s', characterName='%s', nameSearch='%s', minLevel=%d, maxLevel=%d, equipsTo='%s'. Found %d items.", itemType, itemSubType, characterName, nameSearch, minLevel, maxLevel, equipsTo, len(filteredItems))

		// Calculate total pages
		totalPages := (len(filteredItems) + itemsPerPage - 1) / itemsPerPage
		if totalPages == 0 {
			totalPages = 1 // Ensure at least one page even if no items
		}

		// Apply pagination
		startIndex := (page - 1) * itemsPerPage
		endIndex := startIndex + itemsPerPage
		if startIndex >= len(filteredItems) {
			startIndex = 0 // Reset to first page if page is out of bounds
			endIndex = itemsPerPage
			page = defaultPage
		}
		if endIndex > len(filteredItems) {
			endIndex = len(filteredItems)
		}
		paginatedItems := filteredItems[startIndex:endIndex]

		templates.Index(paginatedItems, db.GetUniqueItemTypes(allItems.Items), itemType, db.GetUniqueItemSubTypes(allItems.Items), itemSubType, db.GetUniqueCharacterNames(allItems.Items), characterName, minLevel, maxLevel, page, totalPages, len(filteredItems), db.GetUniqueEquipsTo(allItems.Items), equipsTo).Render(context.Background(), w) // Pass uniqueEquipsTo and equipsTo
	}))

	http.HandleFunc("/filter", func(w http.ResponseWriter, r *http.Request) {
		// Acquire read lock before accessing allItems
		itemsMutex.RLock()
		defer itemsMutex.RUnlock()

		itemType := r.URL.Query().Get("itemType")
		if itemType == "" {
			itemType = "All" // Default to "All" if not specified
		}

		itemSubType := r.URL.Query().Get("itemSubType")
		if itemSubType == "" {
			itemSubType = "All" // Default to "All" if not specified
		}

		characterName := r.URL.Query().Get("characterName")
		if characterName == "" {
			characterName = "All" // Default to "All" if not specified
		}

		nameSearch := r.URL.Query().Get("nameSearch")

		minLevelStr := r.URL.Query().Get("minLevel")
		minLevel, err := strconv.Atoi(minLevelStr)
		if err != nil || minLevel < 0 {
			minLevel = 0 // Default minimum level
		}

		maxLevelStr := r.URL.Query().Get("maxLevel")
		maxLevel, err := strconv.Atoi(maxLevelStr)
		if err != nil || maxLevel < 0 {
			maxLevel = 40 // Default maximum level
		}

		// New: EquipsTo filter parameter
		equipsTo := r.URL.Query().Get("equipsTo")
		if equipsTo == "" {
			equipsTo = "All" // Default to "All" if not specified
		}

		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = defaultPage // Default to page 1
		}

		filteredItems := db.FilterItems(allItems.Items, itemType, itemSubType, characterName, nameSearch, minLevel, maxLevel, equipsTo) // Pass equipsTo
		log.Printf("Filtering by itemType='%s', itemSubType='%s', characterName='%s', nameSearch='%s', minLevel=%d, maxLevel=%d, equipsTo='%s', page %d. Found %d items.", itemType, itemSubType, characterName, nameSearch, minLevel, maxLevel, equipsTo, page, len(filteredItems))

		// Calculate total pages
		totalPages := (len(filteredItems) + itemsPerPage - 1) / itemsPerPage
		if totalPages == 0 {
			totalPages = 1 // Ensure at least one page even if no items
		}

		// Apply pagination
		startIndex := (page - 1) * itemsPerPage
		endIndex := startIndex + itemsPerPage
		if startIndex >= len(filteredItems) {
			startIndex = 0 // Reset to first page if page is out of bounds
			endIndex = itemsPerPage
			page = defaultPage
		}
		if endIndex > len(filteredItems) {
			endIndex = len(filteredItems)
		}
		paginatedItems := filteredItems[startIndex:endIndex]

		// Render only the item list part for HTMX
		templates.ItemList(paginatedItems, itemType, itemSubType, characterName, page, totalPages, len(filteredItems), equipsTo).Render(context.Background(), w) // Pass equipsTo
	})

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
