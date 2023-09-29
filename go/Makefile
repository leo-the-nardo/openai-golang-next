createmigration:
	migrate create -ext=sql -dir=sql/migrations -seq init

migrateup:
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/chat_test" -verbose up

migratedown:
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/chat_test" -verbose drop

.PHONY: migrateup createmigration migratedown