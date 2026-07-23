package system

import "testing"

func TestDefaultMenuComponentNamesArePresentAndUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for _, group := range DefaultMenus() {
		for _, menu := range group.Children {
			if menu.Component == nil {
				continue
			}
			if menu.ComponentName == nil || *menu.ComponentName == "" {
				t.Fatalf("menu %q has no component name", menu.Name)
			}
			if _, exists := seen[*menu.ComponentName]; exists {
				t.Fatalf("duplicate component name %q", *menu.ComponentName)
			}
			seen[*menu.ComponentName] = struct{}{}
		}
	}
}
