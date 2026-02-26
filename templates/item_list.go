package templates

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fingon/ddo-trove-ui/db"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components" //nolint:revive,staticcheck
	. "maragu.dev/gomponents/html"       //nolint:revive,staticcheck
)

const (
	paginationInclude = "#nameSearch, #minLevel, #maxLevel, #equipsToFilter"
	paginationClass   = "pagination-button"
	btcSuffix         = " (BTC)"
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

func paginationPath(selectedType, selectedSubType, selectedCharacter string, page int, selectedEquipsTo string) string {
	values := url.Values{}
	values.Set("item_type", selectedType)
	values.Set("item_sub_type", selectedSubType)
	values.Set("character_name", selectedCharacter)
	values.Set("page", strconv.Itoa(page))
	values.Set("equips_to", selectedEquipsTo)
	return fmt.Sprintf("%s?%s", itemsEndpoint, values.Encode())
}

func paginationControls(selectedType, selectedSubType, selectedCharacter string, currentPage, totalPages int, selectedEquipsTo string) g.Node {
	return Div(Class("pagination-controls"),
		g.If(currentPage > 1,
			Button(
				Class(paginationClass),
				Data("hx-get", paginationPath(selectedType, selectedSubType, selectedCharacter, currentPage-1, selectedEquipsTo)),
				Data("hx-target", itemListContainerID),
				Data("hx-swap", hxSwapMode),
				Data("hx-include", paginationInclude),
				g.Text("Previous"),
			),
		),
		generatePageButtons(selectedType, selectedSubType, selectedCharacter, currentPage, totalPages, selectedEquipsTo),
		g.If(currentPage < totalPages,
			Button(
				Class(paginationClass),
				Data("hx-get", paginationPath(selectedType, selectedSubType, selectedCharacter, currentPage+1, selectedEquipsTo)),
				Data("hx-target", itemListContainerID),
				Data("hx-swap", hxSwapMode),
				Data("hx-include", paginationInclude),
				g.Text("Next"),
			),
		),
	)
}

func generatePageButtons(selectedType, selectedSubType, selectedCharacter string, currentPage, totalPages int, selectedEquipsTo string) g.Node {
	var buttons []g.Node
	pageRange := getPageRange(currentPage, totalPages)

	for _, page := range pageRange {
		buttons = append(buttons,
			Button(
				Classes{paginationClass: true, "active": page == currentPage},
				Data("hx-get", paginationPath(selectedType, selectedSubType, selectedCharacter, page, selectedEquipsTo)),
				Data("hx-target", itemListContainerID),
				Data("hx-swap", hxSwapMode),
				Data("hx-include", paginationInclude),
				g.Text(strconv.Itoa(page)),
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
	if item.Binding == db.BindingBoundToCharacter {
		return Div(Class("item-name btc"), g.Text(item.Name+btcSuffix))
	}
	return Div(Class("item-name"), g.Text(item.Name))
}

func itemTooltip(item db.Item) g.Node {
	var content []g.Node

	if item.Binding == db.BindingBoundToCharacter {
		content = append(content, H4(Class("btc"), g.Text(item.Name+btcSuffix)))
	} else {
		content = append(content, H4(g.Text(item.Name)))
	}

	content = append(content,
		labeledText("Type", item.ItemType),
		labeledText("Character", item.CharacterName),
		labeledText("Quantity", strconv.Itoa(item.Quantity)),
		labeledText("Minimum Level", strconv.Itoa(item.MinimumLevel)),
		labeledText("Location", fmt.Sprintf("%s - %s (Tab %d), Row %d, Col %d", item.Container, item.TabName, item.Tab, item.Row, item.Column)),
	)

	if len(item.EquipsTo) > 0 {
		content = append(content, labeledText("Equips To", strings.Join(item.EquipsTo, ", ")))
	}

	if item.Description != "" {
		content = append(content, labeledText("Description", item.Description))
	}

	if item.Clicky != nil {
		content = append(content, labeledText("Clicky", fmt.Sprintf("%s (CL %d)", item.Clicky.SpellName, item.Clicky.CasterLevel)))
	}

	if len(item.AugmentSlots) > 0 {
		content = append(content, P(Strong(g.Text("Augment Slots:"))))
		var slots []g.Node
		for _, slot := range item.AugmentSlots {
			slots = append(slots, Li(g.Text(fmt.Sprintf("%s (%s)", slot.Name, slot.Color))))
		}
		content = append(content, g.El("ul", g.Group(slots)))
	}

	if len(item.Effects) > 0 {
		content = append(content, P(Strong(g.Text("Effects:"))))
		var effects []g.Node
		for _, effect := range item.Effects {
			effects = append(effects, Li(g.Text(fmt.Sprintf("%s: %s", effect.Name, effect.Description))))
		}
		content = append(content, g.El("ul", g.Group(effects)))
	}

	return Div(Class("item-tooltip"), g.Group(content))
}

func labeledText(label, value string) g.Node {
	return P(Strong(g.Text(label+": ")), g.Text(value))
}
