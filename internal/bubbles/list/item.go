package list

import (
	"go.gopad.dev/fuzzysearch/fuzzy"
)

type Item interface {
	Title() string
	Description() string
}

type FilterItem interface {
	Item
	FilterValue() string
}

type modelItem[T Item] struct {
	index int
	item  T
}

func (m modelItem[T]) FilterValue() string {
	if f, ok := any(m.item).(FilterItem); ok {
		return f.FilterValue()
	}
	return m.item.Title()
}

func parseItems[T Item](items []T) []modelItem[T] {
	mItems := make([]modelItem[T], 0, len(items))
	for i, item := range items {
		mItems = append(mItems, modelItem[T]{
			index: i,
			item:  item,
		})
	}
	return mItems
}

func modelItems[T Item](mItems []modelItem[T]) []T {
	items := make([]T, 0, len(mItems))
	for _, mItem := range mItems {
		items = append(items, mItem.item)
	}
	return items
}

func filterItems[T Item](items []modelItem[T], filter string) []modelItem[T] {
	if filter == "" {
		return items
	}

	return fuzzy.FindNormalizedFold(filter, items)
}
