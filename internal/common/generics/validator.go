package generics

import (
	"context"
	"github.com/shopspring/decimal"
	error2 "hideout/internal/pkg/error"
	"time"
)

func (p FromTo[T]) Validate(ctx context.Context) error {
	if !p.Valid() {
		return error2.ErrInvalidParameter
	}

	return nil
}

func (p FromTo[T]) IsZero() bool {
	switch from := any(p.From).(type) {
	case decimal.Decimal:
		{
			to := any(p.To).(decimal.Decimal)
			return from.IsZero() && to.IsZero()
		}
	case uint:
		{
			to := any(p.To).(uint)
			return from == 0 && to == 0
		}

	case time.Time:
		{
			to := any(p.To).(time.Time)
			return from.IsZero() && to.IsZero()
		}
	}

	return false
}

func (p FromTo[T]) Valid() bool {
	switch from := any(p.From).(type) {
	case decimal.Decimal:
		{
			to := any(p.To).(decimal.Decimal)
			if !(from.IsZero() && to.IsZero()) {
				if from.GreaterThan(to) {
					return false
				}
			}

			if to.LessThan(from) {
				return false
			}
		}
	case uint:
		{
			to := any(p.To).(uint)
			if !(from == 0 && to == 0) {
				if from > to {
					return false
				}
			}

			if to < from {
				return false
			}
		}

	case time.Time:
		{
			to := any(p.To).(time.Time)
			if !(from.IsZero() && to.IsZero()) {
				if from.After(to) {
					return false
				}

				if to.Before(from) {
					return false
				}
			}
		}
	}

	return true
}
