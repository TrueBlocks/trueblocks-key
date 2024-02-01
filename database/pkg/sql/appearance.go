package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func InsertAppearance(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH ids AS (
    INSERT INTO %[1]s (address)
    VALUES ($1::varchar(42))
    ON CONFLICT DO NOTHING
    RETURNING id AS address_id
),
present_ids AS (
    SELECT address_id FROM ids
    UNION ALL
    SELECT id AS address_id FROM %[1]s WHERE address = $1 LIMIT 1
)
INSERT INTO %[2]s (address_id, block_number, tx_id)
SELECT present_ids.address_id, $2, $3 FROM present_ids
ON CONFLICT DO NOTHING;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

func SelectAppearances(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT apps.block_number, apps.tx_id
FROM %[1]s
JOIN %[2]s apps ON apps.address_id = id
WHERE address = $1
ORDER BY apps.block_number DESC, tx_id ASC
LIMIT $2
OFFSET $3;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

func SelectAppearancesCount(appearancesTableName string) string {
	return fmt.Sprintf(`
SELECT reltuples::bigint AS estimate
FROM   pg_class
WHERE  oid = 'public.%[1]s'::regclass;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

func SelectAppearancesMaxBlockNumber(appearancesTableName string) string {
	return fmt.Sprintf(`
SELECT max(block_number)
FROM   %[1]s;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

func SelectAppearancesCountForAddress(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT count(*)
FROM %[1]s
JOIN %[2]s apps ON apps.address_id = id
WHERE address = $1;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}
