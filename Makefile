DB_URL=postgresql://root:account@localhost:5432/account_book?sslmode=disable

postgres:
	docker run --name postgres12 --network account-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=account -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root account_book

dropdb:
	docker exec -it postgres12 dropdb account_book

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: test server
