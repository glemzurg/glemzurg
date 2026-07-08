package parser_ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassAssociationMapKey(t *testing.T) {
	tests := []struct {
		name     string
		assoc    *inputClassAssociation
		nameKey  string
		expected string
	}{
		{
			name: "subdomain level",
			assoc: &inputClassAssociation{
				FromClassKey: "book_order",
				ToClassKey:   "book_order_line",
			},
			nameKey:  "has_lines",
			expected: "book_order--book_order_line--has_lines",
		},
		{
			name: "domain level",
			assoc: &inputClassAssociation{
				FromClassKey: "orders/book_order",
				ToClassKey:   "shipping/shipment",
			},
			nameKey:  "ships_via",
			expected: "orders.book_order--shipping.shipment--ships_via",
		},
		{
			name: "model level",
			assoc: &inputClassAssociation{
				FromClassKey: "order_fulfillment/default/book_order_line",
				ToClassKey:   "inventory/default/inventory_item",
			},
			nameKey:  "reserves",
			expected: "order_fulfillment.default.book_order_line--inventory.default.inventory_item--reserves",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classAssociationMapKey(tt.assoc, tt.nameKey)
			assert.Equal(t, tt.expected, got)
		})
	}
}
