package secrets

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

func (rq GetSecretsRQ) Validate(ctx context.Context, Localizer *i18n.Localizer) error {
	errPagination := rq.Pagination.Validate(ctx)
	if errPagination != nil {
		return errors.Wrap(errPagination, "Pagination validation failed")
	}

	for _, orderVal := range rq.Order {
		errOrdering := orderVal.Validate(ctx, Localizer)
		if errOrdering != nil {
			return errors.Wrap(errOrdering, "Order validation failed")
		}
	}

	return nil
}
