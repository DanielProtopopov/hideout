package ordering

import "context"

func (rq *OrderRQ) Validate(ctx context.Context) error {
	if rq.OrderBy != "" {
		// @TODO Validation
	} else {
		rq.OrderBy = "ID"
	}
	return nil
}
