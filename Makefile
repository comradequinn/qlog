.PHONY: *

build :
	@go build

test :
	@go test -v -cover ./...

example :
	@echo "open a second terminal window and run 'make example-requests'. send ctrl+c to stop"
	@echo ""
	-@cd example && go build -o ./.example && ./.example -logfmt
	-rm -rf ./.example

example-requests:
	@curl -i "http://localhost:8080/echo/?data=world"

bench :
	@go test -run=XXX -bench=. -benchmem

escapes :
	@go build -gcflags '-m' 2>&1 | grep escapes

mem-profile :
	@go test -run=XXX -bench=. -benchmem -memprofile=mem.out
	@go tool pprof -http=localhost:8080 mem.out