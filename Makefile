# Copyright 2017 Keith Irwin. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

PACKAGE = github.com/zentrope/webl

DB_PASS = wanheda
DB_USER = blogsvc
DB_NAME = blogdb

DB_CREATE = create database $(DB_NAME) with encoding 'UTF8'
DB_SETUP = create user $(DB_USER) with login password '$(DB_PASS)' ;\
	alter database $(DB_NAME) owner to $(DB_USER) ;\
	create extension if not exists pgcrypto

.PHONY: build-admin build init govendor vendor vendor-check vendor-unused help
.PHONY: db-clean db-init

.DEFAULT_GOAL := help

# To update dependency versions, govendor fetch +vendor

govendor:
	@hash govendor > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -v -u github.com/kardianos/govendor; \
	fi

ricebox:
	@hash rice > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -v -u github.com/GeertJohan/go.rice/rice; \
	fi

vendor: govendor ## Install govendor and sync deps
	govendor sync

vendor-check: ## Verify that vendored packages match git HEAD
	@git diff-index --quiet HEAD vendor/ || (echo "check-vendor target failed: vendored packages out of sync" && echo && git diff vendor/ && exit 1)

run: vendor ## Run the app from source
	go run main.go

vendor-unused: govendor ## Find unused vendored dependencies
	@govendor list +unused

init: vendor ricebox ## Make sure everything is set up properly for dev.
	cd admin ; yarn

build-admin: ## Build the admin client
	cd admin; yarn ; yarn build

build-freebsd: init build-admin ## Build a version for FreeBSD
	cd internal ; rm -f rice-box.go ;  rice -v embed-go
	GOOS=freebsd GOARCH=amd64 go build -o webl

build: init build-admin ## Build webl into a local binary ./webl.
	cd internal ; rm -f rice-box.go ;  rice -v embed-go
	go build -o webl

clean: ## Clean build artifacts.
	rm -f internal/rice-box.go
	rm -rf webl
	rm -rf admin/build

dist-clean: clean ## Clean everything (vendor, node_modules).
	rm -rf vendor/*/
	rm -rf admin/node_modules

db-clean: ## Delete the local dev database
	psql template1 -c "drop database $(DB_NAME)"
	psql template1 -c "drop user $(DB_USER)"

db-init: ## Create a local dev database with default creds
	psql template1 -c "$(DB_CREATE)"
	psql $(DB_NAME) -c "$(DB_SETUP)"

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort
