QN
===

Project is structure as a collection of modules, which can share code but run independently.
They also have different dependencies - this guarantees small lambda (serverless) executable size.

Structure
---------

1. `config` handling configuration file
1. `database` everything database-related
1. `extract` take whole index and convert it to SQL. Swap tables (staging -> live)
1. `query` lambda (serverless function) and a `cmd` to find appearances
1. `scanner` (deprecated) old lambda to perform appearance lookup
1. `watch` (to be renamed to `listen`) gets notified by scraper about new appearances, manages queue and updates live database
