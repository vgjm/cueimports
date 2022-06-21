.PHONY: help
help:			## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: cueimports
cueimports:		## Compile cueimports.
	@go build -o ./bin/cueimports cmd/cueimports/main.go
	@echo saved to ./bin/cueinports
