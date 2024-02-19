package query

type Paginator interface {
	Pagination() (page uint, perPage uint)
}

func validatePagination(p Paginator) error {
	page, perPage := p.Pagination()

	// Validate pagination
	if page < 0 || perPage < 0 {
		return ErrIncorrectPagePerPage
	}

	if page > MaxSafePage || perPage > MaxSafePerPage {
		return ErrIncorrectPagePerPage
	}
	return nil
}
