package db

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLoadItemsFromDir(t *testing.T) {
	dir := t.TempDir()

	charJSON := `{
		"Name": "CharA",
		"Inventory": [
			{"Name":"Sword","ItemType":"Weapon","ItemSubType":"Sword","MinimumLevel":1,"EquipsTo":["Hand"]}
		]
	}`
	accountJSON := `{
		"SharedBank": {
			"Tabs": {
				"0": {
					"Pages": {
						"0": {
							"Items": [
								{"Name":"Ring","ItemType":"Accessory","MinimumLevel":5,"EquipsTo":["Finger"]}
							]
						}
					}
				}
			}
		}
	}`
	invalidJSON := `{"invalid": true`

	assert.NilError(t, os.WriteFile(filepath.Join(dir, "character.json"), []byte(charJSON), 0o600))
	assert.NilError(t, os.WriteFile(filepath.Join(dir, "account.json"), []byte(accountJSON), 0o600))
	assert.NilError(t, os.WriteFile(filepath.Join(dir, "invalid.json"), []byte(invalidJSON), 0o600))
	assert.NilError(t, os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("noop"), 0o600))

	allItems, err := LoadItemsFromDir(dir)
	assert.NilError(t, err)
	assert.Equal(t, len(allItems.Items), 2)

	assert.Equal(t, allItems.Items[0].Name, "Ring")
	assert.Equal(t, allItems.Items[0].CharacterName, "Account (Shared Bank)")
	assert.Equal(t, allItems.Items[1].Name, "Sword")
	assert.Equal(t, allItems.Items[1].CharacterName, "CharA")
}

func TestFilterItems(t *testing.T) {
	items := []Item{
		{
			Name:          "Flaming Sword",
			ItemType:      "Weapon",
			ItemSubType:   "Sword",
			CharacterName: "CharA",
			MinimumLevel:  5,
			Description:   "A burning blade",
			EquipsTo:      []string{"Hand"},
			Effects:       []Effect{{Name: "Fire Lore", Description: "Boosts fire spells"}},
		},
		{
			Name:          "Icy Ring",
			ItemType:      "Accessory",
			ItemSubType:   "Ring",
			CharacterName: "CharB",
			MinimumLevel:  10,
			Description:   "Cold protection",
			EquipsTo:      []string{"Finger"},
			Effects:       []Effect{{Name: "Cold Resist", Description: "Resists cold"}},
		},
		{
			Name:          "Arcane Cloak",
			ItemType:      "Armor",
			ItemSubType:   "Cloak",
			CharacterName: "CharA",
			MinimumLevel:  8,
			Description:   "Spell focus",
			EquipsTo:      []string{"Back"},
			Effects:       []Effect{{Name: "Spell Power", Description: "Arcane bonus"}},
			Clicky:        &Clicky{SpellName: "Teleport", SpellDescription: "Travel quickly"},
		},
	}

	testCases := []struct {
		name         string
		itemType     string
		itemSubType  string
		character    string
		nameSearch   string
		minLevel     int
		maxLevel     int
		equipsTo     string
		expectedSize int
		firstName    string
	}{
		{
			name:         "all",
			itemType:     FilterAll,
			itemSubType:  FilterAll,
			character:    FilterAll,
			nameSearch:   "",
			minLevel:     0,
			maxLevel:     40,
			equipsTo:     FilterAll,
			expectedSize: 3,
			firstName:    "Arcane Cloak",
		},
		{
			name:         "type and character",
			itemType:     "Weapon",
			itemSubType:  FilterAll,
			character:    "CharA",
			nameSearch:   "",
			minLevel:     0,
			maxLevel:     40,
			equipsTo:     FilterAll,
			expectedSize: 1,
			firstName:    "Flaming Sword",
		},
		{
			name:         "full text orders name matches first",
			itemType:     FilterAll,
			itemSubType:  FilterAll,
			character:    FilterAll,
			nameSearch:   "cold",
			minLevel:     0,
			maxLevel:     40,
			equipsTo:     FilterAll,
			expectedSize: 1,
			firstName:    "Icy Ring",
		},
		{
			name:         "equips to",
			itemType:     FilterAll,
			itemSubType:  FilterAll,
			character:    FilterAll,
			nameSearch:   "",
			minLevel:     0,
			maxLevel:     40,
			equipsTo:     "Back",
			expectedSize: 1,
			firstName:    "Arcane Cloak",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := FilterItems(
				items,
				testCase.itemType,
				testCase.itemSubType,
				testCase.character,
				testCase.nameSearch,
				testCase.minLevel,
				testCase.maxLevel,
				testCase.equipsTo,
			)

			assert.Equal(t, len(result), testCase.expectedSize)
			if testCase.expectedSize > 0 {
				assert.Equal(t, result[0].Name, testCase.firstName)
			}
		})
	}
}

func TestGetUniqueHelpers(t *testing.T) {
	items := []Item{
		{ItemType: "Weapon", ItemSubType: "Sword", CharacterName: "CharA", EquipsTo: []string{"Hand", "Finger"}},
		{ItemType: "Weapon", ItemSubType: "Axe", CharacterName: "CharB", EquipsTo: []string{"Hand"}},
		{ItemType: "Armor", ItemSubType: "", CharacterName: "", EquipsTo: []string{"Body"}},
	}

	assert.DeepEqual(t, GetUniqueItemTypes(items), []string{"Armor", "Weapon"})
	assert.DeepEqual(t, GetUniqueItemSubTypes(items), []string{"Axe", "Sword"})
	assert.DeepEqual(t, GetUniqueCharacterNames(items), []string{"CharA", "CharB"})
	assert.DeepEqual(t, GetUniqueEquipsTo(items), []string{"Body", "Finger", "Hand"})
}
