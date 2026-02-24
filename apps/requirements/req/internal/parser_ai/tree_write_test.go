package parser_ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassAssociationFilename(t *testing.T) {
	tests := []struct {
		name     string
		assoc    *inputClassAssociation
		level    AssociationLevel
		expected string
	}{
		{
			name: "subdomain level",
			assoc: &inputClassAssociation{
				FromClassKey: "book_order",
				ToClassKey:   "book_order_line",
				Name:         "order lines",
			},
			level:    AssocLevelSubdomain,
			expected: "book_order--book_order_line--order_lines.assoc.json",
		},
		{
			name: "domain level",
			assoc: &inputClassAssociation{
				FromClassKey: "orders/book_order",
				ToClassKey:   "shipping/shipment",
				Name:         "order shipment",
			},
			level:    AssocLevelDomain,
			expected: "orders.book_order--shipping.shipment--order_shipment.assoc.json",
		},
		{
			name: "model level",
			assoc: &inputClassAssociation{
				FromClassKey: "order_fulfillment/default/book_order_line",
				ToClassKey:   "inventory/default/inventory_item",
				Name:         "order inventory",
			},
			level:    AssocLevelModel,
			expected: "order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory.assoc.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classAssociationFilename(tt.assoc, tt.level)
			assert.Equal(t, tt.expected, got)
		})
	}
}
