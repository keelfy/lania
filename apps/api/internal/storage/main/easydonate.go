package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const findEDProductsByProductIDs = `
SELECT DISTINCT ed_product_id 
FROM ed_products 
WHERE product_id IN ('%s')
`

func (q *queries) FindEDProductsByProductIDs(ctx context.Context, productIDs uuid.UUIDs) ([]int64, error) {
	productIDsStr := make([]string, len(productIDs))
	for i, productID := range productIDs {
		productIDsStr[i] = productID.String()
	}
	query := fmt.Sprintf(findEDProductsByProductIDs, strings.Join(productIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return []int64{}, err
	}
	defer rows.Close()

	edProductIDs := make([]int64, 0)
	for rows.Next() {
		var edProductID int64
		err := rows.Scan(&edProductID)
		if err != nil {
			return []int64{}, err
		}
		edProductIDs = append(edProductIDs, edProductID)
	}
	return edProductIDs, nil
}
