package pricing

import "github.com/phx7/albion-killgot/internal/albion"

// EventItems extracts all equipment and inventory items from the victim.
func EventItems(event albion.Event) []ItemSpec {
	var items []ItemSpec
	eq := event.Victim.Equipment
	for _, slot := range []*albion.Item{
		eq.MainHand, eq.OffHand, eq.Head, eq.Armor, eq.Shoes,
		eq.Bag, eq.Cape, eq.Mount, eq.Potion, eq.Food,
	} {
		if slot != nil && slot.Type != "" {
			items = append(items, ItemSpec{ID: slot.Type, Quality: slot.Quality, Count: 1})
		}
	}
	for _, inv := range event.Victim.Inventory {
		if inv.Type != "" {
			items = append(items, ItemSpec{ID: inv.Type, Quality: inv.Quality, Count: inv.Count})
		}
	}
	return items
}
