package sql

import "fmt"

func CreateTableChunks(tableName string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
    cid varchar(46) unique not null,
	range varchar(47) not null,
    author varchar(50) not null
);
`, tableName)
}
