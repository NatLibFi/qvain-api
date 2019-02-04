package psql

const (
	sqlObjectInsert     = "INSERT INTO objects(owner, family, schema, type, blob) VALUES($1, $2, $3, $4, $5)"
	sqlObjectSelectId   = "SELECT blob FROM objects WHERE id = $1 and owner = $2"
	sqlObjectSelectList = "SELECT blob FROM objects WHERE owner = $1 and family = $2 and schema = $3 and type = $4"
)
