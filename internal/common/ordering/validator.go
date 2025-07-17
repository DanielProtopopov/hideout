package ordering

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (rq *OrderRQ) Validate(ctx context.Context, Localizer *i18n.Localizer) error {
	if rq.OrderBy != "" {
		// @TODO Validation
	} else {
		rq.OrderBy = "ID"
	}
	return nil
}
