include .env
export

migrateup:
	migrate -path db/sqlc/migration -database "$(DB_SOURCE)" -verbose up

migratedown:
	migrate -path db/sqlc/migration -database "$(DB_SOURCE)" -verbose down

new_migration:
	migrate create -ext sql -dir db/sqlc/migration -seq $(name)

sqlc:
	sqlc generate

server:
	go run main.go

test:
	go test -v -cover ./...

uplocal:
	docker-compose -f ./docker-compose.local.yml up -d 

upprod:
	docker-compose up -d 

dlocal:
	docker-compose -f ./docker-compose.local.yml down 

dprod:
	docker-compose down

.PHONY: migrateup migratedown new_migration sqlc server test local prod