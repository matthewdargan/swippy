.PHONY: build clean deploy

build:
	env GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ebay-find-by-keyword ebay-find-by-keyword/main.go
	env GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ebay-find-advanced ebay-find-advanced/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
