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

func SelectAppearancesFirstPage(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH addrs AS (
    SELECT id
    FROM %[1]s
    WHERE address = @address
)
SELECT block_number, tx_id
FROM %[2]s
WHERE block_number <= @lastBlock AND address_id = (SELECT id FROM addrs)
ORDER BY block_number DESC, tx_id DESC
LIMIT @pageSize;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

// SelectAppearancesNextPage needs the last appearance from the current page.
func SelectAppearancesNextPage(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
WITH addrs AS (
    SELECT id
    FROM %[1]s
    WHERE address = @address
)
SELECT block_number, tx_id
FROM %[2]s
WHERE block_number <= @lastBlock AND address_id = (SELECT id FROM addrs) AND (block_number, tx_id) < (@appBlockNumber, @appTransactionIndex)
ORDER BY block_number DESC, tx_id DESC
LIMIT @pageSize;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}

// SelectAppearancesPreviousPage needs the first appearance from the current page.
// It inverts ordering to read the previous page
func SelectAppearancesPreviousPage(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT * FROM (
	WITH addrs AS (
    	SELECT id
    	FROM %[1]s
    	WHERE address = @address
	)
	SELECT block_number, tx_id
	FROM %[2]s
	WHERE block_number <= @lastBlock AND address_id = (SELECT id FROM addrs) AND (block_number, tx_id) > (@appBlockNumber, @appTransactionIndex)
	ORDER BY block_number ASC, tx_id ASC
	LIMIT @pageSize
) AS x ORDER BY block_number DESC, tx_id ASC;
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

func SelectAppearancesCountForAddress(appearancesTableName string, addressesTableName string) string {
	return fmt.Sprintf(`
SELECT count(*)
FROM %[1]s
JOIN %[2]s apps ON apps.address_id = id
WHERE address = $1 AND block_number <= $2;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{addressesTableName}),
		pgx.Identifier.Sanitize(pgx.Identifier{appearancesTableName}),
	)
}
