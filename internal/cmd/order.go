package cmd

import (
	"cmp"
	"slices"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
)

// sortedApps orders command processing without changing the saved config order.
func sortedApps(apps []config.App) []config.App {
	apps = slices.Clone(apps)
	slices.SortStableFunc(apps, func(a, b config.App) int {
		if order := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	return apps
}
