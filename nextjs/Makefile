#proto-loader-gen-types --long=String --enums=String --defaults --oneofs --grpcLib=@grpc/grpc-js --outDir=./src/grpc/rpc ./proto/*.proto /// to makefile

createmigration:
	todo

migrateup:
	todo

migratedown:
	todo

grpc:
	rm -rf ./src/grpc/rpc && proto-loader-gen-types --long=String --enums=String --defaults --oneofs --grpcLib=@grpc/grpc-js --outDir=./src/grpc/rpc ./proto/*.proto

.PHONY: migrateup createmigration migratedown grpc