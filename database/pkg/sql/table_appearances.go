package sql

import "fmt"

func CreateTableAppearances(tableName string, addressesTableName string) string {
	constraintName := tableName + "_appearances_unique"
	return fmt.Sprintf(`
CREATE TABLE %s(
    address_id BIGINT REFERENCES %s(id) ON DELETE RESTRICT,
    block_number INTEGER,
    tx_id INTEGER,
    CONSTRAINT %s UNIQUE(address_id, block_number, tx_id)
);
`, tableName, addressesTableName, constraintName)
}

func CreateAppearancesOrderIndex(tableName string) string {
	indexName := tableName + "_appearances_order"
	return fmt.Sprintf(`
CREATE INDEX %s ON %s (block_number DESC NULLS LAST, tx_id ASC NULLS LAST);
`, indexName, tableName)
}
