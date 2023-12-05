package sql

import "fmt"

func CreateTableAddresses(tableName string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
    id BIGSERIAL UNIQUE,
    address VARCHAR(42) UNIQUE
);
`, tableName)
}
