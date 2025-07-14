package pagination

// Offset calculate offset
func (m Pagination) Offset() uint {
	return (m.Page - 1) * m.PerPage
}

// Limit calculate limit
func (m Pagination) Limit() uint {
	return m.PerPage
}
