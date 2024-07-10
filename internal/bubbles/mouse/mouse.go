package mouse

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/input"
	"github.com/lrstanley/bubblezone"
)

func Matches(msg any, id string, button input.MouseButton, keyMod ...tea.KeyMod) bool {
	var z *zone.ZoneInfo
	if id != "" {
		z = zone.Get(id)
	}

	return MatchesZone(msg, z, button, keyMod...)
}

func MatchesZone(msg any, zone *zone.ZoneInfo, button input.MouseButton, keyMod ...tea.KeyMod) bool {
	mouseMsg, ok := msg.(tea.MouseEvent)
	if !ok {
		return false
	}

	if zone != nil {
		if !zone.InBounds(mouseMsg) {
			return false
		}
	}

	if mouseMsg.Button != button {
		return false
	}

	if len(keyMod) > 0 {
		var allKeyMods tea.KeyMod
		for _, mod := range keyMod {
			allKeyMods |= mod
		}

		if mouseMsg.Mod != allKeyMods {
			return false
		}
	}

	return true
}
