# Simulate Session Test

This is a load test that simulates users performing actions same as the real ones.

## Usage

Create a TOML file in `./scenarios` (the directory content is git-ignored), like this:
```toml
baseUrl = "{{ API URL }}"
duration = "10s"

[scenarios.1]
address = "0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"
perPage = 20
headers.headerKey = "value"

[scenarios.2]
address = "0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"
perPage = 10
headers.headerKey = "value2"
```

Then run:
```bash
go run . --path scenarios/file.toml
```

Scenarios will be ran concurrently, but requests inside them sequentially (to simulate user's workflow).
Scenarios use `pageId` to navigate through dataset.

You need to use `headers.XXX = "YYY"` to set headers needed for authorization.
