package sql

import (
	"strings"
	"testing"
)

// The catalog has to order by the whole cursor key, or products sharing a creation second would repeat or vanish between pages.
func TestFindProductsPageOrdersByCursorKey(t *testing.T) {
	if !strings.Contains(findProductsPage, "ORDER BY products.category DESC, products.created_at DESC, products.id DESC") {
		t.Error("findProductsPage does not order by category, created_at and id descending")
	}
}
