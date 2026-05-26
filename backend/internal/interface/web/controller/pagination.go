package controller

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

func normalizePage(page *int) {
	if *page <= 0 {
		*page = defaultPage
	}
}

func normalizePageSize(pageSize *int) {
	if *pageSize <= 0 {
		*pageSize = defaultPageSize
		return
	}
	if *pageSize > maxPageSize {
		*pageSize = maxPageSize
	}
}

func normalizePagination(page *int, pageSize *int) {
	normalizePage(page)
	normalizePageSize(pageSize)
}
