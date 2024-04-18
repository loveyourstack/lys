
# install CLI (must use this when testing so that latest config is always used)
.PHONY: cli
cli:
	sudo cp lys_config.toml /usr/local/etc
	go install ./internal/cmd/lyscli

# run all tests
.PHONY: tests
tests: 
	go test -race ./...

# (re-)create test database
.PHONY: testdb
testdb: cli
	lyscli createTestDb
