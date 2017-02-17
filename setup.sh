#!/usr/bin/env bash -eux

go fmt ./...
go run cmd/qbg/main.go -type Sample -output misc/fixture/a/model_query.go misc/fixture/a
go run cmd/qbg/main.go -type Sample -output misc/fixture/b/model_query.go misc/fixture/b
go run cmd/qbg/main.go -output misc/fixture/c/model_query.go misc/fixture/c
go run cmd/qbg/main.go -output misc/fixture/d/model_query.go misc/fixture/d
go run cmd/qbg/main.go -output misc/fixture/e/model_query.go misc/fixture/e
go run cmd/qbg/main.go -output misc/fixture/f/model_query.go misc/fixture/f
go run cmd/qbg/main.go -output misc/fixture/g/model_query.go -private misc/fixture/g
