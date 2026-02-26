package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fingon/ddo-trove-ui/db"
	"gotest.tools/v3/assert"
)

func TestParseFilterParams(t *testing.T) {
	app := &App{}
	testCases := []struct {
		name     string
		query    string
		expected FilterParams
	}{
		{
			name:  "defaults",
			query: "",
			expected: FilterParams{
				ItemType:      db.FilterAll,
				ItemSubType:   db.FilterAll,
				CharacterName: db.FilterAll,
				NameSearch:    "",
				EquipsTo:      db.FilterAll,
				MinLevel:      defaultMinLevel,
				MaxLevel:      defaultMaxLevel,
				Page:          defaultPage,
			},
		},
		{
			name:  "all fields",
			query: "?item_type=Weapon&item_sub_type=Sword&character_name=CharA&name_search=fire&equips_to=Hand&min_level=4&max_level=20&page=3",
			expected: FilterParams{
				ItemType:      "Weapon",
				ItemSubType:   "Sword",
				CharacterName: "CharA",
				NameSearch:    "fire",
				EquipsTo:      "Hand",
				MinLevel:      4,
				MaxLevel:      20,
				Page:          3,
			},
		},
		{
			name:  "invalid numeric values",
			query: "?min_level=-1&max_level=nope&page=0",
			expected: FilterParams{
				ItemType:      db.FilterAll,
				ItemSubType:   db.FilterAll,
				CharacterName: db.FilterAll,
				NameSearch:    "",
				EquipsTo:      db.FilterAll,
				MinLevel:      defaultMinLevel,
				MaxLevel:      defaultMaxLevel,
				Page:          defaultPage,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+testCase.query, nil)
			actual := app.parseFilterParams(req)
			assert.DeepEqual(t, actual, testCase.expected)
		})
	}
}

func TestApplyFilterAndPaginate(t *testing.T) {
	app := &App{}
	items := []db.Item{
		{Name: "A", ItemType: "Weapon", ItemSubType: "Sword", CharacterName: "CharA", MinimumLevel: 1},
		{Name: "B", ItemType: "Armor", ItemSubType: "Heavy", CharacterName: "CharB", MinimumLevel: 2},
	}

	testCases := []struct {
		name         string
		params       FilterParams
		expectedPage int
		expectedSize int
	}{
		{
			name: "default page",
			params: FilterParams{
				ItemType:      db.FilterAll,
				ItemSubType:   db.FilterAll,
				CharacterName: db.FilterAll,
				EquipsTo:      db.FilterAll,
				MinLevel:      defaultMinLevel,
				MaxLevel:      defaultMaxLevel,
				Page:          defaultPage,
			},
			expectedPage: 1,
			expectedSize: 2,
		},
		{
			name: "out of range page resets",
			params: FilterParams{
				ItemType:      db.FilterAll,
				ItemSubType:   db.FilterAll,
				CharacterName: db.FilterAll,
				EquipsTo:      db.FilterAll,
				MinLevel:      defaultMinLevel,
				MaxLevel:      defaultMaxLevel,
				Page:          999,
			},
			expectedPage: 1,
			expectedSize: 2,
		},
		{
			name: "filtered to zero",
			params: FilterParams{
				ItemType:      "Accessory",
				ItemSubType:   db.FilterAll,
				CharacterName: db.FilterAll,
				EquipsTo:      db.FilterAll,
				MinLevel:      defaultMinLevel,
				MaxLevel:      defaultMaxLevel,
				Page:          defaultPage,
			},
			expectedPage: 1,
			expectedSize: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := app.applyFilterAndPaginate(items, testCase.params)
			assert.Equal(t, result.Page, testCase.expectedPage)
			assert.Equal(t, len(result.Items), testCase.expectedSize)
		})
	}
}

func TestNeedsReload(t *testing.T) {
	now := time.Now()
	older := now.Add(-time.Minute)

	testCases := []struct {
		name     string
		oldTimes map[string]time.Time
		newTimes map[string]time.Time
		expect   bool
	}{
		{
			name:     "same",
			oldTimes: map[string]time.Time{"a.json": now},
			newTimes: map[string]time.Time{"a.json": now},
			expect:   false,
		},
		{
			name:     "new file",
			oldTimes: map[string]time.Time{"a.json": now},
			newTimes: map[string]time.Time{"a.json": now, "b.json": now},
			expect:   true,
		},
		{
			name:     "updated file",
			oldTimes: map[string]time.Time{"a.json": older},
			newTimes: map[string]time.Time{"a.json": now},
			expect:   true,
		},
		{
			name:     "deleted file",
			oldTimes: map[string]time.Time{"a.json": now, "b.json": now},
			newTimes: map[string]time.Time{"a.json": now},
			expect:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, needsReload(testCase.oldTimes, testCase.newTimes), testCase.expect)
		})
	}
}

func TestParseConfig(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		cfg, err := parseConfig([]string{"--port", "9090", "--reload-interval", "2m", "-v", "./data"})
		assert.NilError(t, err)
		assert.Equal(t, cfg.Port, 9090)
		assert.Equal(t, cfg.ReloadInterval, 2*time.Minute)
		assert.Equal(t, cfg.Verbose, true)
		assert.DeepEqual(t, cfg.Dirs, []string{"./data"})
	})

	t.Run("env overrides", func(t *testing.T) {
		t.Setenv("DDO_TROVE_PORT", "7070")
		t.Setenv("DDO_TROVE_RELOAD_INTERVAL", "3m")
		t.Setenv("DDO_TROVE_VERBOSE", "true")
		cfg, err := parseConfig([]string{"./data"})
		assert.NilError(t, err)
		assert.Equal(t, cfg.Port, 7070)
		assert.Equal(t, cfg.ReloadInterval, 3*time.Minute)
		assert.Equal(t, cfg.Verbose, true)
	})

	t.Run("missing directories", func(t *testing.T) {
		_, err := parseConfig([]string{})
		assert.Assert(t, err != nil)
	})
}

func TestRoutesAndHandlers(t *testing.T) {
	items := []db.Item{{
		Name:          "Flaming Sword",
		ItemType:      "Weapon",
		ItemSubType:   "Sword",
		CharacterName: "CharA",
		MinimumLevel:  5,
		Quantity:      1,
		EquipsTo:      []string{"Hand"},
	}}
	app := &App{
		cfg:            Config{Port: defaultPort, ReloadInterval: defaultReload, Dirs: []string{"."}},
		allItems:       &db.AllItems{Items: items},
		fileModTimes:   map[string]time.Time{},
		itemTypes:      db.GetUniqueItemTypes(items),
		itemSubTypes:   db.GetUniqueItemSubTypes(items),
		characterNames: db.GetUniqueCharacterNames(items),
		equipsToValues: db.GetUniqueEquipsTo(items),
	}

	handler := app.routes()

	t.Run("index route", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(recorder, request)
		assert.Equal(t, recorder.Code, 200)
		assert.Assert(t, strings.Contains(recorder.Body.String(), "DDO Trove Item Browser"))
		assert.Assert(t, strings.Contains(recorder.Body.String(), "Flaming Sword"))
	})

	t.Run("items route", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/items?item_type=Weapon&page=1", nil)
		handler.ServeHTTP(recorder, request)
		assert.Equal(t, recorder.Code, 200)
		assert.Assert(t, strings.Contains(recorder.Body.String(), "Found 1 items."))
		assert.Assert(t, strings.Contains(recorder.Body.String(), "Flaming Sword"))
	})
}
