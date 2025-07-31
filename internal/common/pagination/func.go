package pagination

import "math"

// CountPages для получения общего количества страниц
func CountPages(total uint, paging PaginationRQ) PaginationRS {
	if paging.PerPage == 0 {
		return PaginationRS{PerPage: 0, Pages: 0, Total: 0}
	}

	return PaginationRS{
		PerPage: paging.PerPage,
		Pages:   uint(math.Ceil(float64(total) / float64(paging.PerPage))),
		Total:   total,
	}
}

func Paginate(length int, skip int, size int) (int, int) {
	if skip > length {
		skip = length
	}

	end := skip + size
	if end > length {
		end = length
	}

	return skip, end
}
