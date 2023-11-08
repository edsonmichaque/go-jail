##
# Project Title
#
# @file
# @version 0.2



# end

.PHONY: proto

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
