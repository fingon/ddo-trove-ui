package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const FilterAll = "All"

// Root structure for character-specific bank/inventory files
type CharacterData struct {
	CharacterID         int64      `json:"CharacterId"`
	Name                string     `json:"Name"`
	LastUpdated         *time.Time `json:"LastUpdated"` // Use pointer for optional/null
	PersonalBank        *Bank      `json:"PersonalBank"`
	ReincarnationBank   *Bank      `json:"ReincarnationBank"`
	Inventory           []Item     `json:"Inventory"` // Inventory is a direct list of items
	Server              string     `json:"Server"`
	SubscriptionKeyHash string     `json:"SubscriptionKeyHash"`
	SubscriptionAlias   *string    `json:"SubscriptionAlias"` // Use pointer for optional/null
	UsedCapacity        int        `json:"UsedCapacity"`
	MaxCapacity         int        `json:"MaxCapacity"`
}

// Root structure for account-wide shared bank file
type AccountData struct {
	SharedBank          *Bank   `json:"SharedBank"`
	CraftingBank        *Bank   `json:"CraftingBank"`
	Server              string  `json:"Server"`
	SubscriptionKeyHash string  `json:"SubscriptionKeyHash"`
	SubscriptionAlias   *string `json:"SubscriptionAlias"`
	UsedCapacity        int     `json:"UsedCapacity"`
	MaxCapacity         int     `json:"MaxCapacity"`
}

// Generic Bank structure (Personal, Reincarnation, Shared, Crafting)
type Bank struct {
	BankType int            `json:"BankType"`
	Tabs     map[string]Tab `json:"Tabs"`
}

type Tab struct {
	ID    int             `json:"Id"`
	Name  string          `json:"Name"`
	Index int             `json:"Index"`
	Pages map[string]Page `json:"Pages"`
}

type Page struct {
	Items []Item `json:"Items"`
}

// Item structure (common for all containers)
type Item struct {
	OwnerID              int64         `json:"OwnerId"`
	CharacterName        string        `json:"CharacterName"`
	ItemID               int64         `json:"ItemId"`
	Container            string        `json:"Container"`
	Tab                  int           `json:"Tab"`
	TabName              string        `json:"TabName"`
	Row                  int           `json:"Row"`
	Column               int           `json:"Column"`
	Quantity             int           `json:"Quantity"`
	WeenieID             int64         `json:"WeenieId,omitempty"`
	Charges              int           `json:"Charges,omitempty"`
	MaxCharges           int           `json:"MaxCharges,omitempty"`
	TreasureType         string        `json:"TreasureType"`
	Name                 string        `json:"Name"`
	Description          string        `json:"Description"`
	MinimumLevel         int           `json:"MinimumLevel"`
	Binding              string        `json:"Binding,omitempty"`
	ItemType             string        `json:"ItemType"`
	BaseValueCopper      int           `json:"BaseValueCopper"`
	Hardness             int64         `json:"Hardness,omitempty"`
	EquipsToFlags        int           `json:"EquipsToFlags"`
	EquipsTo             []string      `json:"EquipsTo"`
	IconSource           string        `json:"IconSource"`
	Clicky               *Clicky       `json:"Clicky,omitempty"`
	AugmentSlots         []AugmentSlot `json:"AugmentSlots"`
	Proficiency          string        `json:"Proficiency,omitempty"`
	WeaponType           string        `json:"WeaponType,omitempty"`
	ItemSubType          string        `json:"ItemSubType,omitempty"`
	ArmorType            string        `json:"ArmorType,omitempty"`
	Effects              []Effect      `json:"Effects"`
	Hover                string        `json:"Hover"`
	SetBonus1Name        string        `json:"SetBonus1Name,omitempty"`
	SetBonus1Description []string      `json:"SetBonus1Description,omitempty"`
	MinorArtifact        bool          `json:"MinorArtifact,omitempty"`
}

type Clicky struct {
	SpellName        string   `json:"SpellName"`
	SpellDescription string   `json:"SpellDescription"`
	CasterLevel      int      `json:"CasterLevel"`
	ValidTargets     []string `json:"ValidTargets"`
}

type AugmentSlot struct {
	Name  string `json:"Name"`
	Color string `json:"Color"`
}

type Effect struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// AllItems holds all parsed items from all JSON files
type AllItems struct {
	Items []Item
}

// LoadItemsFromDir reads all JSON files from a directory and parses them into Item structs.
func LoadItemsFromDir(dirPath string) (*AllItems, error) {
	var allItems AllItems

	files, err := os.ReadDir(dirPath) // Changed from ioutil.ReadDir
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Failed to read file %s: %v\n", filePath, err)
			continue
		}

		// Try to unmarshal as CharacterData first
		var charData CharacterData
		if err := json.Unmarshal(data, &charData); err == nil {
			// Assign CharacterName to items from PersonalBank, ReincarnationBank, and Inventory
			if charData.PersonalBank != nil {
				for _, tab := range charData.PersonalBank.Tabs {
					for _, page := range tab.Pages {
						for i := range page.Items {
							page.Items[i].CharacterName = charData.Name
						}
						allItems.Items = append(allItems.Items, page.Items...)
					}
				}
			}
			if charData.ReincarnationBank != nil {
				for _, tab := range charData.ReincarnationBank.Tabs {
					for _, page := range tab.Pages {
						for i := range page.Items {
							page.Items[i].CharacterName = charData.Name
						}
						allItems.Items = append(allItems.Items, page.Items...)
					}
				}
			}
			if charData.Inventory != nil {
				for i := range charData.Inventory {
					charData.Inventory[i].CharacterName = charData.Name
				}
				allItems.Items = append(allItems.Items, charData.Inventory...)
			}
			continue // Successfully parsed as CharacterData
		}

		// If not CharacterData, try to unmarshal as AccountData
		var accountData AccountData
		if err := json.Unmarshal(data, &accountData); err == nil {
			// Assign a default "Account" character name to shared bank items
			if accountData.SharedBank != nil {
				for _, tab := range accountData.SharedBank.Tabs {
					for _, page := range tab.Pages {
						for i := range page.Items {
							page.Items[i].CharacterName = "Account (Shared Bank)"
						}
						allItems.Items = append(allItems.Items, page.Items...)
					}
				}
			}
			if accountData.CraftingBank != nil {
				for _, tab := range accountData.CraftingBank.Tabs {
					for _, page := range tab.Pages {
						for i := range page.Items {
							page.Items[i].CharacterName = "Account (Crafting Bank)"
						}
						allItems.Items = append(allItems.Items, page.Items...)
					}
				}
			}
			continue // Successfully parsed as AccountData
		}

		fmt.Printf("Warning: Could not unmarshal %s as either CharacterData or AccountData. Error: %v\n", filePath, err)
	}

	return &allItems, nil // Return fileModTimes
}

// FilterItems filters a slice of items by item type, item sub type, character name, name search, minimum level range, and equips to.
func FilterItems(items []Item, itemType, itemSubType, characterName, nameSearch string, minLevel, maxLevel int, equipsTo string) []Item {
	var nameMatches []Item
	var effectMatches []Item

	searchLower := strings.ToLower(nameSearch)

	for _, item := range items {
		matchItemType := (itemType == "" || itemType == FilterAll || item.ItemType == itemType)
		matchItemSubType := (itemSubType == "" || itemSubType == FilterAll || item.ItemSubType == itemSubType)
		matchCharacterName := (characterName == "" || characterName == FilterAll || item.CharacterName == characterName)
		matchMinLevel := (item.MinimumLevel >= minLevel && item.MinimumLevel <= maxLevel)

		// EquipsTo filter
		matchEquipsTo := (equipsTo == "" || equipsTo == FilterAll)
		if !matchEquipsTo { // Only check if a specific filter is selected
			for _, eq := range item.EquipsTo {
				if eq == equipsTo {
					matchEquipsTo = true
					break
				}
			}
		}

		// Skip if basic filters don't match
		if !matchItemType || !matchItemSubType || !matchCharacterName || !matchMinLevel || !matchEquipsTo {
			continue
		}

		// Full text search logic
		if nameSearch == "" {
			nameMatches = append(nameMatches, item)
		} else {
			// Check name match first
			nameMatch := strings.Contains(strings.ToLower(item.Name), searchLower)

			// Check effects match
			effectMatch := false
			for _, effect := range item.Effects {
				if strings.Contains(strings.ToLower(effect.Name), searchLower) ||
					strings.Contains(strings.ToLower(effect.Description), searchLower) {
					effectMatch = true
					break
				}
			}

			// Also check description and clicky
			descMatch := strings.Contains(strings.ToLower(item.Description), searchLower)
			clickyMatch := false
			if item.Clicky != nil {
				clickyMatch = strings.Contains(strings.ToLower(item.Clicky.SpellName), searchLower) ||
					strings.Contains(strings.ToLower(item.Clicky.SpellDescription), searchLower)
			}

			if nameMatch {
				nameMatches = append(nameMatches, item)
			} else if effectMatch || descMatch || clickyMatch {
				effectMatches = append(effectMatches, item)
			}
		}
	}

	// Sort each group by name
	sort.Slice(nameMatches, func(i, j int) bool {
		return nameMatches[i].Name < nameMatches[j].Name
	})
	sort.Slice(effectMatches, func(i, j int) bool {
		return effectMatches[i].Name < effectMatches[j].Name
	})

	// Combine results with name matches first
	return append(nameMatches, effectMatches...)
}

// GetUniqueItemTypes extracts all unique item types from a slice of items.
func GetUniqueItemTypes(items []Item) []string {
	seen := make(map[string]bool)
	var types []string
	for _, item := range items {
		if !seen[item.ItemType] {
			seen[item.ItemType] = true
			types = append(types, item.ItemType)
		}
	}
	sort.Strings(types)
	return types
}

// GetUniqueCharacterNames extracts all unique character names from a slice of items.
func GetUniqueCharacterNames(items []Item) []string {
	seen := make(map[string]bool)
	var names []string
	for _, item := range items {
		if item.CharacterName != "" && !seen[item.CharacterName] {
			seen[item.CharacterName] = true
			names = append(names, item.CharacterName)
		}
	}
	sort.Strings(names)
	return names
}

// GetUniqueItemSubTypes extracts all unique item sub types from a slice of items.
func GetUniqueItemSubTypes(items []Item) []string {
	seen := make(map[string]bool)
	var subTypes []string
	for _, item := range items {
		if item.ItemSubType != "" && !seen[item.ItemSubType] {
			seen[item.ItemSubType] = true
			subTypes = append(subTypes, item.ItemSubType)
		}
	}
	sort.Strings(subTypes)
	return subTypes
}

// GetUniqueEquipsTo extracts all unique "EquipsTo" values from a slice of items.
func GetUniqueEquipsTo(items []Item) []string {
	seen := make(map[string]bool)
	var equipsToValues []string
	for _, item := range items {
		for _, eq := range item.EquipsTo {
			if eq != "" && !seen[eq] {
				seen[eq] = true
				equipsToValues = append(equipsToValues, eq)
			}
		}
	}
	sort.Strings(equipsToValues)
	return equipsToValues
}
