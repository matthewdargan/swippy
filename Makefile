.PHONY: build zip clean init plan apply

build:
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find-by-category/bootstrap find-by-category/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find-by-keywords/bootstrap find-by-keywords/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find-advanced/bootstrap find-advanced/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find-by-product/bootstrap find-by-product/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find-in-ebay-stores/bootstrap find-in-ebay-stores/main.go

zip: build
	zip -j bin/find-by-category.zip bin/find-by-category/bootstrap
	zip -j bin/find-by-keywords.zip bin/find-by-keywords/bootstrap
	zip -j bin/find-advanced.zip bin/find-advanced/bootstrap
	zip -j bin/find-by-product.zip bin/find-by-product/bootstrap
	zip -j bin/find-in-ebay-stores.zip bin/find-in-ebay-stores/bootstrap

clean:
	rm -rf ./bin

init:
	tofu init

plan: clean zip init
	tofu plan

apply: clean zip init
	tofu apply
