package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fingon/ddo-trove-ui/db"
	"github.com/fingon/ddo-trove-ui/templates"
)

const (
	itemsPerPage      = 100
	defaultPage       = 1
	defaultMinLevel   = 0
	defaultMaxLevel   = 40
	defaultPort       = 8080
	defaultReload     = time.Minute
	itemsPath         = "/items"
	staticPathPrefix  = "/static/"
	localhostTemplate = "http://localhost:%d"
)

type Config struct {
	Port           int           `default:"8080" env:"DDO_TROVE_PORT" help:"HTTP port."`
	ReloadInterval time.Duration `default:"1m" env:"DDO_TROVE_RELOAD_INTERVAL" help:"Polling interval for data reload." name:"reload-interval"`
	Verbose        bool          `env:"DDO_TROVE_VERBOSE" help:"Enable debug logging." short:"v"`
	Dirs           []string      `arg:"" help:"Input directories with Trove JSON files." name:"dirs"`
}

func (c Config) Validate() (err error) {
	if len(c.Dirs) == 0 {
		return errors.New("at least one input directory is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	if c.ReloadInterval <= 0 {
		return fmt.Errorf("reload interval must be positive, got %s", c.ReloadInterval)
	}
	return nil
}

type FilterParams struct {
	ItemType      string
	ItemSubType   string
	CharacterName string
	NameSearch    string
	EquipsTo      string
	MinLevel      int
	MaxLevel      int
	Page          int
}

type PaginationResult struct {
	Items      []db.Item
	Page       int
	TotalPages int
	TotalCount int
}

type App struct {
	cfg Config

	mu           sync.RWMutex
	allItems     *db.AllItems
	fileModTimes map[string]time.Time

	itemTypes      []string
	itemSubTypes   []string
	characterNames []string
	equipsToValues []string
}

func parseConfig(args []string) (cfg Config, err error) {
	parser, err := kong.New(
		&cfg,
		kong.Name("ddo-trove-ui"),
		kong.Description("Web UI for browsing DDO Trove item data."),
		kong.UsageOnError(),
	)
	if err != nil {
		return cfg, fmt.Errorf("create parser: %w", err)
	}
	if _, err = parser.Parse(args); err != nil {
		return cfg, fmt.Errorf("parse arguments: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

func configureLogging(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

func newApp(cfg Config) (app *App, err error) {
	items, err := loadAndAggregateItems(cfg.Dirs)
	if err != nil {
		return nil, fmt.Errorf("initial load: %w", err)
	}

	modTimes := collectFileModTimes(cfg.Dirs)

	app = &App{
		cfg:            cfg,
		allItems:       items,
		fileModTimes:   modTimes,
		itemTypes:      db.GetUniqueItemTypes(items.Items),
		itemSubTypes:   db.GetUniqueItemSubTypes(items.Items),
		characterNames: db.GetUniqueCharacterNames(items.Items),
		equipsToValues: db.GetUniqueEquipsTo(items.Items),
	}

	slog.Info("initial load complete", "items", len(items.Items), "dirs", len(cfg.Dirs))
	return app, nil
}

func (a *App) parseFilterParams(r *http.Request) FilterParams {
	query := r.URL.Query()
	params := FilterParams{
		ItemType:      query.Get("item_type"),
		ItemSubType:   query.Get("item_sub_type"),
		CharacterName: query.Get("character_name"),
		NameSearch:    query.Get("name_search"),
		EquipsTo:      query.Get("equips_to"),
		MinLevel:      defaultMinLevel,
		MaxLevel:      defaultMaxLevel,
		Page:          defaultPage,
	}

	if params.ItemType == "" {
		params.ItemType = db.FilterAll
	}
	if params.ItemSubType == "" {
		params.ItemSubType = db.FilterAll
	}
	if params.CharacterName == "" {
		params.CharacterName = db.FilterAll
	}
	if params.EquipsTo == "" {
		params.EquipsTo = db.FilterAll
	}

	if minLevelStr := query.Get("min_level"); minLevelStr != "" {
		if minLevel, convErr := strconv.Atoi(minLevelStr); convErr == nil && minLevel >= 0 {
			params.MinLevel = minLevel
		}
	}
	if maxLevelStr := query.Get("max_level"); maxLevelStr != "" {
		if maxLevel, convErr := strconv.Atoi(maxLevelStr); convErr == nil && maxLevel >= 0 {
			params.MaxLevel = maxLevel
		}
	}
	if pageStr := query.Get("page"); pageStr != "" {
		if page, convErr := strconv.Atoi(pageStr); convErr == nil && page >= 1 {
			params.Page = page
		}
	}

	return params
}

func (a *App) applyFilterAndPaginate(items []db.Item, params FilterParams) PaginationResult {
	filteredItems := db.FilterItems(
		items,
		params.ItemType,
		params.ItemSubType,
		params.CharacterName,
		params.NameSearch,
		params.MinLevel,
		params.MaxLevel,
		params.EquipsTo,
	)

	totalCount := len(filteredItems)
	totalPages := (totalCount + itemsPerPage - 1) / itemsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	page := params.Page
	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage
	if startIndex >= totalCount {
		startIndex = 0
		endIndex = itemsPerPage
		page = defaultPage
	}
	if endIndex > totalCount {
		endIndex = totalCount
	}

	return PaginationResult{
		Items:      filteredItems[startIndex:endIndex],
		Page:       page,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}
}

func loadAndAggregateItems(dirPaths []string) (combinedAllItems *db.AllItems, err error) {
	combinedAllItems = &db.AllItems{}
	for _, dirPath := range dirPaths {
		absPath, absErr := filepath.Abs(dirPath)
		if absErr != nil {
			return nil, fmt.Errorf("resolve absolute path for %q: %w", dirPath, absErr)
		}

		info, statErr := os.Stat(absPath)
		if errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("input directory does not exist, skipping", "path", absPath)
			continue
		}
		if statErr != nil {
			return nil, fmt.Errorf("stat input path %q: %w", absPath, statErr)
		}
		if !info.IsDir() {
			slog.Warn("input path is not a directory, skipping", "path", absPath)
			continue
		}

		dirItems, loadErr := db.LoadItemsFromDir(absPath)
		if loadErr != nil {
			slog.Error("failed loading items from directory", "path", absPath, "err", loadErr)
			continue
		}
		combinedAllItems.Items = append(combinedAllItems.Items, dirItems.Items...)
	}
	return combinedAllItems, nil
}

func collectFileModTimes(dirPaths []string) map[string]time.Time {
	currentFileModTimes := make(map[string]time.Time)
	for _, dirPath := range dirPaths {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			slog.Warn("failed to read directory while collecting mod times", "path", dirPath, "err", err)
			continue
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			filePath := filepath.Join(dirPath, file.Name())
			info, statErr := os.Stat(filePath)
			if statErr != nil {
				slog.Warn("failed to stat file while collecting mod times", "path", filePath, "err", statErr)
				continue
			}
			currentFileModTimes[filePath] = info.ModTime()
		}
	}
	return currentFileModTimes
}

func needsReload(oldTimes, newTimes map[string]time.Time) bool {
	if len(oldTimes) != len(newTimes) {
		return true
	}
	for path, oldTime := range oldTimes {
		newTime, exists := newTimes[path]
		if !exists || newTime.After(oldTime) {
			return true
		}
	}
	for path, newTime := range newTimes {
		oldTime, exists := oldTimes[path]
		if !exists || newTime.After(oldTime) {
			return true
		}
	}
	return false
}

func (a *App) monitorAndReloadItems() {
	newModTimes := collectFileModTimes(a.cfg.Dirs)

	a.mu.RLock()
	oldModTimes := make(map[string]time.Time, len(a.fileModTimes))
	for key, value := range a.fileModTimes {
		oldModTimes[key] = value
	}
	a.mu.RUnlock()

	if !needsReload(oldModTimes, newModTimes) {
		return
	}

	slog.Info("detected data change, reloading")
	newAllItems, err := loadAndAggregateItems(a.cfg.Dirs)
	if err != nil {
		slog.Error("failed to reload items", "err", err)
		return
	}

	a.mu.Lock()
	a.allItems = newAllItems
	a.fileModTimes = newModTimes
	a.itemTypes = db.GetUniqueItemTypes(newAllItems.Items)
	a.itemSubTypes = db.GetUniqueItemSubTypes(newAllItems.Items)
	a.characterNames = db.GetUniqueCharacterNames(newAllItems.Items)
	a.equipsToValues = db.GetUniqueEquipsTo(newAllItems.Items)
	a.mu.Unlock()

	slog.Info("reload complete", "items", len(newAllItems.Items))
}

func (a *App) startMonitor(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.ReloadInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.monitorAndReloadItems()
			}
		}
	}()
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	items := a.allItems.Items
	itemTypes := append([]string(nil), a.itemTypes...)
	itemSubTypes := append([]string(nil), a.itemSubTypes...)
	characterNames := append([]string(nil), a.characterNames...)
	equipsToValues := append([]string(nil), a.equipsToValues...)
	a.mu.RUnlock()

	params := a.parseFilterParams(r)
	result := a.applyFilterAndPaginate(items, params)

	slog.Info("render index",
		"item_type", params.ItemType,
		"item_sub_type", params.ItemSubType,
		"character_name", params.CharacterName,
		"name_search", params.NameSearch,
		"min_level", params.MinLevel,
		"max_level", params.MaxLevel,
		"equips_to", params.EquipsTo,
		"page", result.Page,
		"count", result.TotalCount,
	)

	if err := templates.Index(
		result.Items,
		itemTypes,
		params.ItemType,
		itemSubTypes,
		params.ItemSubType,
		characterNames,
		params.CharacterName,
		params.MinLevel,
		params.MaxLevel,
		result.Page,
		result.TotalPages,
		result.TotalCount,
		equipsToValues,
		params.EquipsTo,
	).Render(w); err != nil {
		slog.Error("render index failed", "err", err)
		http.Error(w, "failed to render index", http.StatusInternalServerError)
	}
}

func (a *App) handleItems(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	items := a.allItems.Items
	a.mu.RUnlock()

	params := a.parseFilterParams(r)
	result := a.applyFilterAndPaginate(items, params)

	slog.Debug("render items",
		"item_type", params.ItemType,
		"item_sub_type", params.ItemSubType,
		"character_name", params.CharacterName,
		"name_search", params.NameSearch,
		"min_level", params.MinLevel,
		"max_level", params.MaxLevel,
		"equips_to", params.EquipsTo,
		"page", result.Page,
		"count", result.TotalCount,
	)

	if err := templates.ItemList(
		result.Items,
		params.ItemType,
		params.ItemSubType,
		params.CharacterName,
		result.Page,
		result.TotalPages,
		result.TotalCount,
		params.EquipsTo,
	).Render(w); err != nil {
		slog.Error("render items failed", "err", err)
		http.Error(w, "failed to render items", http.StatusInternalServerError)
	}
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(staticPathPrefix, http.StripPrefix(staticPathPrefix, http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc(itemsPath, a.handleItems)
	return mux
}

func run(args []string) (err error) {
	cfg, err := parseConfig(args)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	configureLogging(cfg.Verbose)
	app, err := newApp(cfg)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.startMonitor(ctx)

	address := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("server starting", "url", fmt.Sprintf(localhostTemplate, cfg.Port), "address", address)

	server := &http.Server{
		Addr:    address,
		Handler: app.routes(),
	}

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("application failed", "err", err)
		os.Exit(1)
	}
}
