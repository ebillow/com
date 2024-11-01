protoc26.1 -I=. --go_out=../pb --go-grpc_out=../pb/ --go-grpc_opt=paths=source_relative common.proto configs.proto

@echo "Finish build proto files."