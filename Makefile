
# install CLI (must use this when testing so that latest config is always used)
.PHONY: cli
cli:
	sudo cp lys_config.toml /usr/local/etc
	go install ./internal/cmd/lyscli

PROTO_FILES := $(shell find . -type f -name '*.proto' -not -path './.git/*')

# generate protobuf code (using @ to suppress Make's command echo, since list of files can be long. stdout and stderr are unaffected)
.PHONY: genpb
genpb:
	@protoc --go_out=. --go_opt=module=github.com/loveyourstack/lys \
		--go-grpc_out=. --go-grpc_opt=module=github.com/loveyourstack/lys \
		$(PROTO_FILES)

# run all tests
.PHONY: tests
tests: genpb
	go test -race ./...

# (re-)create test database
.PHONY: testdb
testdb: cli
	lyscli createTestDb
