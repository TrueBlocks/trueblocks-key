mktestuser
==========

Creates `--count` number of test endpoints and stores them in a file.
When used with `--remove` flag, it reads endpoints from the file and removes them.

It can also generate testing scenarios for our session simulator.

Usage
-----

```bash
# Create 10 test endpoints:
AWS_PROFILE=staging-or-production-profile go run . --file endpoints.txt --count 10

# Remove the above endpoints
AWS_PROFILE=staging-or-production-profile go run . --file endpoints.txt --remove

# Generate testing scenarios for ../tests/simulate_session
go run . --file endpoints.txt --scenario scenarios.txt
```