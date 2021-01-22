SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o frp-adapter main.go
	@docker build -t frp-adapter .
	@rm frp-adapter

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag frp-adapter 10.1.11.196/library/frp-adapter/frp-adapter41-$(REV)
	@docker push 10.1.11.196/library/frp-adapter/frp-adapter41-$(REV)
