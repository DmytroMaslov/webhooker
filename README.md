### Webhooker
---
Webhooker receives events from payment system and streams them to clients

### How to run
repo contains `Makefile` to simplify local testing:
- `make up-env` - start environment (postgresdb in docker file)
- `make run` - run server
- `down-env` - stop environment
- `make clean` - remove tmp files

### How to test
File `scripts/webhookerdb_init/init.sql` contains predefined orders.