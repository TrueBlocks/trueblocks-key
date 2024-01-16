package sql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func InsertChunk(chunksTableName string) string {
	return fmt.Sprintf(`
INSERT INTO %[1]s (cid, range, author)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;
`,
		pgx.Identifier.Sanitize(pgx.Identifier{chunksTableName}),
	)
}

func SelectDuplicatedChunks(chunksTableName string) string {
	return fmt.Sprintf(`
SELECT range FROM %[1]s
GROUP BY range
HAVING count(*) > 1
`,
		pgx.Identifier.Sanitize(pgx.Identifier{chunksTableName}),
	)
}

func CountChunks(chunksTableName string) string {
	return fmt.Sprintf(`
SELECT count(*) FROM %[1]s
`,
		pgx.Identifier.Sanitize(pgx.Identifier{chunksTableName}),
	)
}
