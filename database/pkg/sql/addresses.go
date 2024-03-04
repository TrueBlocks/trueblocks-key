package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func SelectAddressesInTx(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH apps AS (
    SELECT address_id
    FROM %[1]s
    WHERE block_number = @blockNumber AND tx_id = @transactionIndex
    ORDER BY block_number DESC, tx_id ASC
)
SELECT addrs.address
FROM apps
JOIN %[2]s addrs ON addrs.id = apps.address_id
ORDER BY address;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
	)
}
