package templates

import (
	"strconv"

	"github.com/fingon/ddo-trove-ui/db"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

func Index(items []db.Item, itemTypes []string, selectedType string, itemSubTypes []string, selectedSubType string, characterNames []string, selectedCharacter string, minLevel, maxLevel, currentPage, totalPages, totalFilteredItemsCount int, uniqueEquipsTo []string, selectedEquipsTo string) g.Node {
	return Layout("DDO Trove UI",
		H1(g.Text("DDO Trove Item Browser")),
		Div(Class("filter-controls"),
			Div(Class("filter-row"),
				Label(For("itemTypeFilter"), g.Text("Filter by Item Type:")),
				Select(
					ID("itemTypeFilter"), Name("itemType"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "change"),
					Data("hx-include", "#itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
					selectedOption(db.FilterAll, selectedType),
					g.Group(g.Map(itemTypes, func(itemType string) g.Node { //nolint:unconvert
						return selectedOption(itemType, selectedType)
					})),
				),
				Label(For("itemSubTypeFilter"), g.Text("Item Sub Type:")),
				Select(
					ID("itemSubTypeFilter"), Name("itemSubType"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "change"),
					Data("hx-include", "#itemTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
					selectedOption(db.FilterAll, selectedSubType),
					g.Group(g.Map(itemSubTypes, func(itemSubType string) g.Node { //nolint:unconvert
						return selectedOption(itemSubType, selectedSubType)
					})),
				),
				Label(For("characterFilter"), g.Text("Character:")),
				Select(
					ID("characterFilter"), Name("characterName"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "change"),
					Data("hx-include", "#itemTypeFilter, #itemSubTypeFilter, #nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
					selectedOption(db.FilterAll, selectedCharacter),
					g.Group(g.Map(characterNames, func(charName string) g.Node { //nolint:unconvert
						return selectedOption(charName, selectedCharacter)
					})),
				),
			),
			Div(Class("filter-row"),
				Label(For("equipsToFilter"), g.Text("Equips To:")),
				Select(
					ID("equipsToFilter"), Name("equipsTo"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "change"),
					Data("hx-include", "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #maxLevel"),
					selectedOption(db.FilterAll, selectedEquipsTo),
					g.Group(g.Map(uniqueEquipsTo, func(equipsTo string) g.Node { //nolint:unconvert
						return selectedOption(equipsTo, selectedEquipsTo)
					})),
				),
			),
			Div(Class("filter-row"),
				Label(For("minLevel"), g.Text("Min Level:")),
				Input(Type("number"), ID("minLevel"), Name("minLevel"), Value(strconv.Itoa(minLevel)), Min("0"), Max("40"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "input changed delay:500ms"),
					Data("hx-include", "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #maxLevel, #equipsToFilter"),
				),
				Label(For("maxLevel"), g.Text("Max Level:")),
				Input(Type("number"), ID("maxLevel"), Name("maxLevel"), Value(strconv.Itoa(maxLevel)), Min("0"), Max("40"),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "input changed delay:500ms"),
					Data("hx-include", "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #nameSearch, #minLevel, #equipsToFilter"),
				),
				Label(For("nameSearch"), g.Text("Full Text Search:")),
				Input(Type("text"), ID("nameSearch"), Name("nameSearch"), Placeholder("Search names, effects, descriptions..."),
					Data("hx-get", "/filter"),
					Data("hx-target", "#item-list-container"),
					Data("hx-swap", "innerHTML"),
					Data("hx-trigger", "input changed delay:500ms"),
					Data("hx-include", "#itemTypeFilter, #itemSubTypeFilter, #characterFilter, #minLevel, #maxLevel, #equipsToFilter"),
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
