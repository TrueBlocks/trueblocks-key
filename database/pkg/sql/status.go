package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func SelectStatus(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT max(block_number) as lastIndexedBlock
FROM %[1]s;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}
