createmigration:
	migrate create -ext=sql -dir=sql/migrations -seq init

migrateup:
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/chat_test" -verbose up

migratedown:
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/chat_test" -verbose drop

grpc:
	protoc --go_out=. --go-grpc_out=. --experimental_allow_proto3_optional proto/chat.proto

.PHONY: migrateup createmigration migratedown grpc