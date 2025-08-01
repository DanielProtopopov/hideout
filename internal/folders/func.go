package folders

import (
	"fmt"
	"gorm.io/gorm"
	"hideout/internal/common/model"
	"slices"
	"sort"
	"strings"
)

func (params ListFolderParams) DatabaseFilter(TableName string, Query *gorm.DB) *gorm.DB {
	if len(params.IDs) != 0 {
		Query = Query.Where(TableName+".id IN (?)", params.IDs)
	}
	if len(params.UIDs) != 0 {
		Query = Query.Where(TableName+".uid IN (?)", params.UIDs)
	}

	if params.Pagination.Page != 0 {
		Query = Query.Offset(int(params.Pagination.Offset()))
	}
	if params.Pagination.PerPage != 0 {
		Query = Query.Limit(int(params.Pagination.Limit()))
	}

	if !params.CreatedAt.IsZero() {
		Query = Query.Where(TableName+".created_at BETWEEN ? AND ?", params.CreatedAt.From.UTC(), params.CreatedAt.To.UTC())
	}

	if !params.UpdatedAt.IsZero() {
		Query = Query.Where(TableName+".updated_at BETWEEN ? AND ?", params.UpdatedAt.From.UTC(), params.UpdatedAt.To.UTC())
	}

	if params.Deleted == model.Yes {
		Query = Query.Unscoped().Where(TableName + ".deleted_at IS NOT NULL")
		if !params.DeletedAt.IsZero() {
			Query = Query.Where(TableName+".deleted_at BETWEEN ? AND ?", params.DeletedAt.From.UTC(), params.DeletedAt.To.UTC())
		}
	} else if params.Deleted == model.No {
		Query = Query.Unscoped().Where(TableName + ".deleted_at IS NULL")
	} else if params.Deleted == model.YesOrNo {
		Query = Query.Unscoped()
	}

	return Query
}

func (params ListFolderParams) DatabaseOrder(TableName string, Query *gorm.DB, OrderMap map[string]string) *gorm.DB {
	var results []string
	for _, order := range params.Order {
		orderDirectionVal := "desc"
		if order.Order {
			orderDirectionVal = "asc"
		}
		orderColumn, orderColumnExists := OrderMap[order.OrderBy]
		if orderColumnExists {
			results = append(results, fmt.Sprintf("%s.%s %s", TableName, orderColumn, orderDirectionVal))
		}
	}

	return Query.Order(strings.Join(results, ", "))
}

func (params ListFolderParams) Apply(data map[string][]*Folder) (results map[string][]*Folder) {
	if len(params.IDs) != 0 {
		idResults := make(map[string][]*Folder)
		for folderVal, foldersEntry := range data {
			for _, folder := range foldersEntry {
				if slices.Index(params.IDs, folder.ID) != -1 {
					idResults[folderVal] = append(idResults[folderVal], folder)
				}
			}
		}
		results = idResults
	} else {
		results = data
	}

	if len(params.UIDs) != 0 {
		uidResults := make(map[string][]*Folder)
		for folderVal, foldersEntry := range data {
			for _, folder := range foldersEntry {
				if slices.Index(params.UIDs, folder.UID) != -1 {
					uidResults[folderVal] = append(uidResults[folderVal], folder)
				}
			}
		}
		results = uidResults
	}

	return results
}

type lessFunc func(p1, p2 *Folder) bool

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(folders []*Folder) {
	ms.folders = folders
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.folders)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.folders[i], ms.folders[j] = ms.folders[j], ms.folders[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that discriminates between
// the two items (one is less than the other). Note that it can call the
// less functions twice per call. We could change the functions to return
// -1, 0, 1 and reduce the number of calls for greater efficiency: an
// exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.folders[i], ms.folders[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}
