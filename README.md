# account-book-api

## Docs
- [ER diagram](./docs/er.md)

## How to run

### Setup
``` sh
# 1. start postgres docker
make postgres
# 2. create database
make createdb
# 3. migration
make migrateup
```

### Start server
``` sh
make server
```
