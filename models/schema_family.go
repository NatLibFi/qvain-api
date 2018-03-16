
package models


/*
 * type DatasetModel struct {
 *	Name     string `json:"name"`
 *	baseType int    `json:"family"`
 * }
 */

//func (model *DatasetModel) BaseType() *DatasetType {}

var (
	schemaFamilies = []*SchemaFamily{
		&SchemaFamily{0, "zero", false, []string{}},
		&SchemaFamily{1, "open", true, nil},
		&SchemaFamily{2, "metax", true, []string{"metadata", "contracts", "files"}},
	}
	
	_idMap = make(map[int]*SchemaFamily)
)


func init() {
	for _, fam := range schemaFamilies {
		_idMap[fam.Id] = fam
	}
}


func LookupFamily(id int) (fam *SchemaFamily, found bool) {
	fam, found = _idMap[id]
	return
}


type SchemaFamily struct {
	Id          int
	FamilyName  string
	isVisible   bool
	publicPaths []string `json:"public"`
}


// IsPathPublic returns a boolean indicating if the dataset's subkey can be shown via API.
// A nil path list means no restrictions.
func (fam *SchemaFamily) IsPathPublic(p string) bool {
	if fam.publicPaths == nil {
		return true
	}
	return contains(fam.publicPaths, p)
}


// contains does a simple linear string search.
// Slices are faster than maps for small collections (< 5 elements); as we don't expect to exceed 5 sub-paths, don't create a map.
func contains(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

