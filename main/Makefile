SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms-http-server main.go
	@docker build -t gopherlv/sms-http-server .
	@rm sms-http-server

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/sms-http-server gopherlv/sms_service:sms-http-server-$(REV)
	@docker push gopherlv/sms_service:sms-http-server-$(REV)
