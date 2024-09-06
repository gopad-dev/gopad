package mouse

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/bubblezone"
)

func Matches(msg tea.MouseMsg, id string, button tea.MouseButton, keyMod ...tea.KeyMod) bool {
	var z *zone.ZoneInfo
	if id != "" {
		z = zone.Get(id)
	}

	return MatchesZone(msg, z, button, keyMod...)
}

func MatchesZone(msg tea.MouseMsg, zone *zone.ZoneInfo, button tea.MouseButton, keyMod ...tea.KeyMod) bool {
	if zone == nil {
		return false
	}

	if !zone.InBounds(msg) {
		return false
	}

	mouse := msg.Mouse()

	if mouse.Button != button {
		return false
	}

	if len(keyMod) > 0 {
		var allKeyMods tea.KeyMod
		for _, mod := range keyMod {
			allKeyMods |= mod
		}

		if mouse.Mod != allKeyMods {
			return false
		}
	}

	return true
}
