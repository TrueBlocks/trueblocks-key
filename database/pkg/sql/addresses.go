package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func SelectAddressesInTx(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT addrs.address
FROM %[1]s
JOIN %[2]s addrs ON addrs.id = address_id
WHERE block_number = $1 AND tx_id = $2
LIMIT $3
OFFSET $4;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
	)
}
