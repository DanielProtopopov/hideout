package paths

import (
	"slices"
	"sort"
)

func (params ListPathParams) Apply(data map[string][]*Path) (results map[string][]*Path) {
	if len(params.IDs) != 0 {
		idResults := make(map[string][]*Path)
		for pathVal, pathsEntry := range data {
			for _, path := range pathsEntry {
				if slices.Index(params.IDs, path.ID) != -1 {
					idResults[pathVal] = append(idResults[pathVal], path)
				}
			}
		}
		results = idResults
	} else {
		results = data
	}

	if len(params.UIDs) != 0 {
		uidResults := make(map[string][]*Path)
		for pathVal, pathsEntry := range data {
			for _, path := range pathsEntry {
				if slices.Index(params.UIDs, path.UID) != -1 {
					uidResults[pathVal] = append(uidResults[pathVal], path)
				}
			}
		}
		results = uidResults
	}

	return results
}

type lessFunc func(p1, p2 *Path) bool

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(paths []*Path) {
	ms.paths = paths
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
	return len(ms.paths)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.paths[i], ms.paths[j] = ms.paths[j], ms.paths[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that discriminates between
// the two items (one is less than the other). Note that it can call the
// less functions twice per call. We could change the functions to return
// -1, 0, 1 and reduce the number of calls for greater efficiency: an
// exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.paths[i], ms.paths[j]
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
