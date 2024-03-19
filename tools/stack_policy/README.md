# stack_policy

This tool sets stack policy so that it is not possible to remove or replace
the database.

The idea is to prevent accidental database removal.
If we ever need to replace the database, we will need to set stack policy to
one that allows `Update:*` action on all resources (`Resource:*`) first.

Apart from setting the policy, the tool can also retrieve it and list deployed stacks.

In order for it to work, you have to set the correct user with `AWS_PROFILE` env variable, e.g.
```bash
AWS_PROFILE=key-deployer-production stack_policy --list
```
It must be the user that is deploying the **stack**, not **stack set**, so for example user called
_Production_ in our AWS organization.
