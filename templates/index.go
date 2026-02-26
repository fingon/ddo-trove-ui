package templates

import (
	"strconv"

	"github.com/fingon/ddo-trove-ui/db"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

const (
	itemListContainerID = "#item-list-container"
	itemsEndpoint       = "/items"
	hxSwapMode          = "innerHTML"
	changeTrigger       = "change"
	inputTrigger        = "input changed delay:500ms"

	includeTypeFilter      = "#itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"
	includeSubTypeFilter   = "#itemTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"
	includeCharacterFilter = "#itemTypeFilter, #itemSubTypeFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"
	includeEquipsToFilter  = "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel"
	includeMinLevel        = "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #maxLevel, #equipsToFilter"
	includeMaxLevel        = "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #equipsToFilter"
	includeNameSearch      = "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #minLevel, #maxLevel, #equipsToFilter"
)

func Index(items []db.Item, itemTypes []string, selectedType string, itemSubTypes []string, selectedSubType string, characterNames []string, selectedCharacter string, minLevel, maxLevel, currentPage, totalPages, totalFilteredItemsCount int, uniqueEquipsTo []string, selectedEquipsTo string) g.Node {
	return Layout("DDO Trove UI",
		H1(g.Text("DDO Trove Item Browser")),
		Div(Class("filter-controls"),
			Div(Class("filter-row"),
				Label(For("itemTypeFilter"), g.Text("Filter by Item Type:")),
				Select(
					ID("itemTypeFilter"), Name("item_type"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", changeTrigger),
					Data("hx-include", includeTypeFilter),
					selectedOption(db.FilterAll, selectedType),
					g.Group(g.Map(itemTypes, func(itemType string) g.Node { //nolint:unconvert
						return selectedOption(itemType, selectedType)
					})),
				),
				Label(For("itemSubTypeFilter"), g.Text("Item Sub Type:")),
				Select(
					ID("itemSubTypeFilter"), Name("item_sub_type"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", changeTrigger),
					Data("hx-include", includeSubTypeFilter),
					selectedOption(db.FilterAll, selectedSubType),
					g.Group(g.Map(itemSubTypes, func(itemSubType string) g.Node { //nolint:unconvert
						return selectedOption(itemSubType, selectedSubType)
					})),
				),
				Label(For("characterFilter"), g.Text("Character:")),
				Select(
					ID("characterFilter"), Name("character_name"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", changeTrigger),
					Data("hx-include", includeCharacterFilter),
					selectedOption(db.FilterAll, selectedCharacter),
					g.Group(g.Map(characterNames, func(charName string) g.Node { //nolint:unconvert
						return selectedOption(charName, selectedCharacter)
					})),
				),
			),
			Div(Class("filter-row"),
				Label(For("equipsToFilter"), g.Text("Equips To:")),
				Select(
					ID("equipsToFilter"), Name("equips_to"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", changeTrigger),
					Data("hx-include", includeEquipsToFilter),
					selectedOption(db.FilterAll, selectedEquipsTo),
					g.Group(g.Map(uniqueEquipsTo, func(equipsTo string) g.Node { //nolint:unconvert
						return selectedOption(equipsTo, selectedEquipsTo)
					})),
				),
			),
			Div(Class("filter-row"),
				Label(For("minLevel"), g.Text("Min Level:")),
				Input(Type("number"), ID("minLevel"), Name("min_level"), Value(strconv.Itoa(minLevel)), Min("0"), Max("40"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", inputTrigger),
					Data("hx-include", includeMinLevel),
				),
				Label(For("maxLevel"), g.Text("Max Level:")),
				Input(Type("number"), ID("maxLevel"), Name("max_level"), Value(strconv.Itoa(maxLevel)), Min("0"), Max("40"),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", inputTrigger),
					Data("hx-include", includeMaxLevel),
				),
				Label(For("nameSearch"), g.Text("Full Text Search:")),
				Input(Type("text"), ID("nameSearch"), Name("name_search"), Placeholder("Search names, effects, descriptions..."),
					Data("hx-get", itemsEndpoint),
					Data("hx-target", itemListContainerID),
					Data("hx-swap", hxSwapMode),
					Data("hx-trigger", inputTrigger),
					Data("hx-include", includeNameSearch),
				),
			),
		),
		Div(ID("item-list-container"), Data("hx-preserve", "true"),
			ItemList(items, selectedType, selectedSubType, selectedCharacter, currentPage, totalPages, totalFilteredItemsCount, selectedEquipsTo),
		),
	)
}

func selectedOption(value, selected string) g.Node {
	if value == selected {
		return Option(Value(value), g.Text(value), Selected())
	}
	return Option(Value(value), g.Text(value))
}
