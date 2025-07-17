package secrets

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (rq GetSecretsRQ) Validate(ctx context.Context, Localizer *i18n.Localizer) error {
	errPagination := rq.Pagination.Validate(ctx)
	if errPagination != nil {
		return errPagination
	}

	for _, orderVal := range rq.Ordering {
		errOrdering := orderVal.Validate(ctx, Localizer)
		if errOrdering != nil {
			return errOrdering
		}
	}

	return nil
}
