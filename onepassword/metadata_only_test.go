package onepassword

import (
	"context"
	"reflect"
	"testing"
)

// The Konfirmity build must never read item contents, so every column has to come from the
// row its table's List hydrate streams.
func TestNoColumnHydratesBeyondTheListCall(t *testing.T) {
	ctx := context.Background()
	for name, table := range Plugin(ctx).TableMap {
		var listHydrate uintptr
		if table.List != nil && table.List.Hydrate != nil {
			listHydrate = reflect.ValueOf(table.List.Hydrate).Pointer()
		}
		for _, col := range table.Columns {
			if col.Hydrate == nil {
				continue
			}
			if reflect.ValueOf(col.Hydrate).Pointer() != listHydrate {
				t.Errorf("%s.%s has a hydrate function other than the List hydrate", name, col.Name)
			}
		}
	}
}

func TestItemTableHasNoSecretColumnsOrGetConfig(t *testing.T) {
	table := Plugin(context.Background()).TableMap["onepassword_item"]
	if table == nil {
		t.Fatal("onepassword_item is missing")
	}
	if table.Get != nil {
		t.Error("onepassword_item must not have a Get config, it would fetch the full item")
	}
	columns := map[string]bool{}
	for _, col := range table.Columns {
		columns[col.Name] = true
	}
	for _, removed := range []string{"password", "fields", "sections", "files", "username"} {
		if columns[removed] {
			t.Errorf("onepassword_item still has the %q column", removed)
		}
	}
	for _, kept := range []string{"id", "vault_id", "title", "category", "tags", "urls", "created_at", "updated_at", "state"} {
		if !columns[kept] {
			t.Errorf("onepassword_item lost the %q column", kept)
		}
	}
}
