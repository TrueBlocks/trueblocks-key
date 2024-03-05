package query

type Limiter interface {
	Limit() (perPage uint)
}

func validateLimit(p Limiter) error {
	perPage := p.Limit()

	// Validate perPage
	if perPage < 0 {
		return ErrIncorrectPerPage
	}

	if perPage > MaxSafePerPage {
		return ErrIncorrectPerPage
	}

	return nil
}
