package pagination

import "context"

func (rq *Pagination) Validate(ctx context.Context) error {
	if rq.Page != 0 {
		// @TODO Validation
	} else {
		rq.Page = DefaultPage
	}

	if rq.PerPage != 0 {
		// @TODO Validation
	} else {
		rq.PerPage = DefaultPerOnPage
	}
	return nil
}
