package templates

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fingon/ddo-trove-ui/db"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components" //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html"       //nolint:revive,staticcheck
)

func ItemList(items []db.Item, selectedType, selectedSubType, selectedCharacter string, currentPage, totalPages, totalFilteredItemsCount int, selectedEquipsTo string) g.Node {
	return g.Group([]g.Node{
		paginationControls(selectedType, selectedSubType, selectedCharacter, currentPage, totalPages, selectedEquipsTo),
		Div(Class("pagination-controls")),
		P(Class("item-count"), g.Text(fmt.Sprintf("Found %d items.", totalFilteredItemsCount))),
		Div(Class("item-list"),
			g.If(len(items) == 0,
				P(g.Text("No items found matching the selected criteria.")),
			),
			g.Group(g.Map(items, renderItem)), //nolint:unconvert
		),
		paginationControls(selectedType, selectedSubType, selectedCharacter, currentPage, totalPages, selectedEquipsTo),
	})
}

func paginationControls(selectedType, selectedSubType, selectedCharacter string, currentPage, totalPages int, selectedEquipsTo string) g.Node {
	return Div(Class("pagination-controls"),
		g.If(currentPage > 1,
			Button(
				Class("pagination-button"),
				Data("hx-get", fmt.Sprintf("/filter?itemType=%s&itemSubType=%s&characterName=%s&page=%d&equipsTo=%s", selectedType, selectedSubType, selectedCharacter, currentPage-1, selectedEquipsTo)),
				Data("hx-target", "#item-list-container"),
				Data("hx-swap", "innerHTML"),
				Data("hx-include", "#nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
				g.Text("Previous"),
			),
		),
		generatePageButtons(selectedType, selectedSubType, selectedCharacter, currentPage, totalPages, selectedEquipsTo),
		g.If(currentPage < totalPages,
			Button(
				Class("pagination-button"),
				Data("hx-get", fmt.Sprintf("/filter?itemType=%s&itemSubType=%s&characterName=%s&page=%d&equipsTo=%s", selectedType, selectedSubType, selectedCharacter, currentPage+1, selectedEquipsTo)),
				Data("hx-target", "#item-list-container"),
				Data("hx-swap", "innerHTML"),
				Data("hx-include", "#nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
				g.Text("Next"),
			),
		),
	)
}

func generatePageButtons(selectedType, selectedSubType, selectedCharacter string, currentPage, totalPages int, selectedEquipsTo string) g.Node {
	var buttons []g.Node

	pageRange := getPageRange(currentPage, totalPages)

	for _, i := range pageRange {
		buttons = append(buttons,
			Button(
				Classes{"pagination-button": true, "active": i == currentPage},
				Data("hx-get", fmt.Sprintf("/filter?itemType=%s&itemSubType=%s&characterName=%s&page=%d&equipsTo=%s", selectedType, selectedSubType, selectedCharacter, i, selectedEquipsTo)),
				Data("hx-target", "#item-list-container"),
				Data("hx-swap", "innerHTML"),
				Data("hx-include", "#nameSearch, #minLevel, #maxLevel, #equipsToFilter"),
				g.Text(strconv.Itoa(i)),
			),
		)
	}

	return g.Group(buttons)
}

//nolint:gocritic
func getPageRange(currentPage, totalPages int) []int {
	var pages []int

	if totalPages <= 10 {
		for i := 1; i <= totalPages; i++ {
			pages = append(pages, i)
		}
	} else if currentPage <= 6 {
		for i := 1; i <= 10; i++ {
			pages = append(pages, i)
		}
	} else if currentPage >= totalPages-5 {
		for i := totalPages - 9; i <= totalPages; i++ {
			pages = append(pages, i)
		}
	} else {
		for i := currentPage - 4; i <= currentPage+5; i++ {
			pages = append(pages, i)
		}
	}

	return pages
}

func renderItem(item db.Item) g.Node {
	return Div(Class("item-row"),
		g.If(item.IconSource != "",
			Img(Src(item.IconSource), Alt("Item Icon"), Class("item-icon")),
		),
		itemNameDiv(item),
		Div(Class("item-type"), g.Text(item.ItemType)),
		Div(Class("item-character"), g.Text(item.CharacterName)),
		Div(Class("item-min-level"), g.Text(fmt.Sprintf("Lvl: %d", item.MinimumLevel))),
		Div(Class("item-quantity"), g.Text(fmt.Sprintf("Qty: %d", item.Quantity))),
		Div(Class("item-equips-to"), g.Text("Equips: "+strings.Join(item.EquipsTo, ", "))),
		itemTooltip(item),
	)
}

func itemNameDiv(item db.Item) g.Node {
	if item.Binding == "BoundToCharacter" {
		return Div(Class("item-name btc"), g.Text(item.Name+" (BTC)"))
	}
	return Div(Class("item-name"), g.Text(item.Name))
}

func itemTooltip(item db.Item) g.Node {
	var content []g.Node

	if item.Binding == "BoundToCharacter" {
		content = append(content, H4(Class("btc"), g.Text(item.Name+" (BTC)")))
	} else {
		content = append(content, H4(g.Text(item.Name)))
	}

	content = append(content,
		P(g.Raw("<strong>Type:</strong> "+item.ItemType)),
		P(g.Raw("<strong>Character:</strong> "+item.CharacterName)),
		P(g.Raw(fmt.Sprintf("<strong>Quantity:</strong> %d", item.Quantity))),
		P(g.Raw(fmt.Sprintf("<strong>Minimum Level:</strong> %d", item.MinimumLevel))),
		P(g.Raw(fmt.Sprintf("<strong>Location:</strong> %s - %s (Tab %d), Row %d, Col %d", item.Container, item.TabName, item.Tab, item.Row, item.Column))),
	)

	if len(item.EquipsTo) > 0 {
		content = append(content, P(g.Raw("<strong>Equips To:</strong> "+strings.Join(item.EquipsTo, ", "))))
	}

	if item.Description != "" {
		content = append(content, P(g.Raw("<strong>Description:</strong> "+item.Description)))
	}

	if item.Clicky != nil {
		content = append(content, P(g.Raw(fmt.Sprintf("<strong>Clicky:</strong> %s (CL %d)", item.Clicky.SpellName, item.Clicky.CasterLevel))))
	}

	if len(item.AugmentSlots) > 0 {
		content = append(content, P(g.Raw("<strong>Augment Slots:</strong>")))
		var slots []g.Node
		for _, slot := range item.AugmentSlots {
			slots = append(slots, Li(g.Text(fmt.Sprintf("%s (%s)", slot.Name, slot.Color))))
		}
		content = append(content, g.El("ul", g.Group(slots)))
	}

	if len(item.Effects) > 0 {
		content = append(content, P(g.Raw("<strong>Effects:</strong>")))
		var effects []g.Node
		for _, effect := range item.Effects {
			effects = append(effects, Li(g.Text(fmt.Sprintf("%s: %s", effect.Name, effect.Description))))
		}
		content = append(content, g.El("ul", g.Group(effects)))
	}

	return Div(Class("item-tooltip"), g.Group(content))
}
