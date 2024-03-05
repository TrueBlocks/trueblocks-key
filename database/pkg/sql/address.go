package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// SelectAddressesIn functions don't set the order, because it was too heave on CPU

func SelectAddressesInTx(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH apps AS (
    SELECT address_id
    FROM %[1]s
    WHERE block_number = @blockNumber AND tx_id = @transactionIndex
    ORDER BY block_number DESC, tx_id ASC
)
SELECT DISTINCT addrs.address
FROM apps
JOIN %[2]s addrs ON addrs.id = apps.address_id;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
	)
}

func SelectAddressesInBlock(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH apps AS (
    SELECT DISTINCT address_id
    FROM %[1]s
    WHERE block_number = @blockNumber
)
SELECT addrs.address
FROM apps
JOIN %[2]s addrs ON addrs.id = apps.address_id;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
	)
}
