package db

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	FilterAll               = "All"
	BindingBoundToCharacter = "BoundToCharacter"
	jsonFileSuffix          = ".json"
	accountSharedBankName   = "Account (Shared Bank)"
	accountCraftingBankName = "Account (Crafting Bank)"
)

type CharacterData struct {
	CharacterID         int64      `json:"CharacterId"`
	Name                string     `json:"Name"`
	LastUpdated         *time.Time `json:"LastUpdated"`
	PersonalBank        *Bank      `json:"PersonalBank"`
	ReincarnationBank   *Bank      `json:"ReincarnationBank"`
	Inventory           []Item     `json:"Inventory"`
	Server              string     `json:"Server"`
	SubscriptionKeyHash string     `json:"SubscriptionKeyHash"`
	SubscriptionAlias   *string    `json:"SubscriptionAlias"`
	UsedCapacity        int        `json:"UsedCapacity"`
	MaxCapacity         int        `json:"MaxCapacity"`
}

type AccountData struct {
	SharedBank          *Bank   `json:"SharedBank"`
	CraftingBank        *Bank   `json:"CraftingBank"`
	Server              string  `json:"Server"`
	SubscriptionKeyHash string  `json:"SubscriptionKeyHash"`
	SubscriptionAlias   *string `json:"SubscriptionAlias"`
	UsedCapacity        int     `json:"UsedCapacity"`
	MaxCapacity         int     `json:"MaxCapacity"`
}

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

type AllItems struct {
	Items []Item
}

func LoadItemsFromDir(dirPath string) (allItems *AllItems, err error) {
	allItems = &AllItems{}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", dirPath, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), jsonFileSuffix) {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			slog.Warn("failed to read JSON file", "path", filePath, "err", readErr)
			continue
		}

		var charData CharacterData
		if unmarshalErr := json.Unmarshal(data, &charData); unmarshalErr == nil {
			if hasCharacterPayload(charData) {
				appendItemsFromBank(&allItems.Items, charData.PersonalBank, charData.Name)
				appendItemsFromBank(&allItems.Items, charData.ReincarnationBank, charData.Name)
				appendItemsWithCharacter(&allItems.Items, charData.Inventory, charData.Name)
				continue
			}
		}

		var accountData AccountData
		if unmarshalErr := json.Unmarshal(data, &accountData); unmarshalErr == nil {
			if hasAccountPayload(accountData) {
				appendItemsFromBank(&allItems.Items, accountData.SharedBank, accountSharedBankName)
				appendItemsFromBank(&allItems.Items, accountData.CraftingBank, accountCraftingBankName)
				continue
			}
		}

		slog.Warn("failed to unmarshal file as character/account data", "path", filePath)
	}

	return allItems, nil
}

func hasCharacterPayload(value CharacterData) bool {
	return value.Name != "" || value.PersonalBank != nil || value.ReincarnationBank != nil || len(value.Inventory) > 0
}

func hasAccountPayload(value AccountData) bool {
	return value.SharedBank != nil || value.CraftingBank != nil
}

func appendItemsFromBank(dst *[]Item, bank *Bank, characterName string) {
	if bank == nil {
		return
	}
	for _, tab := range bank.Tabs {
		for _, page := range tab.Pages {
			appendItemsWithCharacter(dst, page.Items, characterName)
		}
	}
}

func appendItemsWithCharacter(dst *[]Item, source []Item, characterName string) {
	if len(source) == 0 {
		return
	}
	items := make([]Item, len(source))
	copy(items, source)
	for index := range items {
		items[index].CharacterName = characterName
	}
	*dst = append(*dst, items...)
}

func FilterItems(items []Item, itemType, itemSubType, characterName, nameSearch string, minLevel, maxLevel int, equipsTo string) []Item {
	var nameMatches []Item
	var effectMatches []Item

	searchLower := strings.ToLower(nameSearch)

	for _, item := range items {
		matchItemType := itemType == "" || itemType == FilterAll || item.ItemType == itemType
		matchItemSubType := itemSubType == "" || itemSubType == FilterAll || item.ItemSubType == itemSubType
		matchCharacterName := characterName == "" || characterName == FilterAll || item.CharacterName == characterName
		matchMinLevel := item.MinimumLevel >= minLevel && item.MinimumLevel <= maxLevel

		matchEquipsTo := equipsTo == "" || equipsTo == FilterAll
		if !matchEquipsTo {
			for _, eq := range item.EquipsTo {
				if eq == equipsTo {
					matchEquipsTo = true
					break
				}
			}
		}

		if !matchItemType || !matchItemSubType || !matchCharacterName || !matchMinLevel || !matchEquipsTo {
			continue
		}

		if nameSearch == "" {
			nameMatches = append(nameMatches, item)
			continue
		}

		nameMatch := strings.Contains(strings.ToLower(item.Name), searchLower)
		effectMatch := false
		for _, effect := range item.Effects {
			if strings.Contains(strings.ToLower(effect.Name), searchLower) ||
				strings.Contains(strings.ToLower(effect.Description), searchLower) {
				effectMatch = true
				break
			}
		}

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

	sort.Slice(nameMatches, func(i, j int) bool {
		return nameMatches[i].Name < nameMatches[j].Name
	})
	sort.Slice(effectMatches, func(i, j int) bool {
		return effectMatches[i].Name < effectMatches[j].Name
	})

	return append(nameMatches, effectMatches...)
}

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
