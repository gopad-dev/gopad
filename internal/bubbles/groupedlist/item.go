package groupedlist

import (
	"go.gopad.dev/fuzzysearch/fuzzy"
)

type Item interface {
	Title() string
	Items() []Item
}

type FilterItem interface {
	Item
	FilterValue() string
}

func filterItems(filter string, item Item) bool {
	if filter == "" {
		return true
	}

	filterValue := item.Title()
	if filterItem, ok := item.(FilterItem); ok {
		filterValue = filterItem.FilterValue()
	}
	if fuzzy.MatchNormalizedFold(filter, filterValue) {
		return true
	}

	for _, i := range item.Items() {
		if filterItems(filter, i) {
			return true
		}
	}

	return false
}
