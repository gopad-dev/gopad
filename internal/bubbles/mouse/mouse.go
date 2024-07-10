package mouse

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/input"
	"github.com/lrstanley/bubblezone"
)

func Matches(msg tea.MouseEvent, id string, button input.MouseButton, keyMod ...tea.KeyMod) bool {
	var z *zone.ZoneInfo
	if id != "" {
		z = zone.Get(id)
	}

	return MatchesZone(msg, z, button, keyMod...)
}

func MatchesZone(msg tea.MouseEvent, zone *zone.ZoneInfo, button input.MouseButton, keyMod ...tea.KeyMod) bool {
	if zone != nil {
		if !zone.InBounds(msg) {
			return false
		}
	}

	if msg.Button != button {
		return false
	}

	if len(keyMod) > 0 {
		var allKeyMods tea.KeyMod
		for _, mod := range keyMod {
			allKeyMods |= mod
		}

		if msg.Mod != allKeyMods {
			return false
		}
	}

	return true
}
