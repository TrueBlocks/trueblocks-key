<img alt="KΞY Powered by TrueBlocks" src="assets/logo_small.png" width="150">

Serverless, SQL based API for serving Unchained Index.

The project is structured as a collection of modules, which can share code but run independently.
They also have different dependencies - this guarantees small lambda (serverless) executable size.

Structure
---------

1. `config` handles configuration files and env variables
1. `database` everything database-related
1. `extract` take whole index and convert it to SQL. Swap tables (staging -> live)
1. `query` lambda (serverless function) and a `cmd` to find appearances
1. `scanner` (deprecated) old lambda to perform appearance lookup
1. `queue` insert to/read from the queue that feeds SQL database
1. `quicknode` QuickNode integration related logic: handling accounts, authorization, provision and healthcheck
1. `test/integration` inntegration tests (they run mocked environment in Docker containers)

How it works
------------

The data can be initially ingested into the queue (and then SQL database) using `extract` tool. Later on, the scraper sends new appearances as notifications to `queue/insert`.

User can sign up using QuickNode Marketplace. This is handled by `quicknode/provision` code.

After signing up, the user can ask for appearances by sending JSON-RPC request to our API. The request is handled by `query/lambda`.

Plans and API keys
------------------

Plans and API keys for the users are defined in SAM template. They include rate limits.
