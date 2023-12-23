.PHONY: build zip clean init plan apply

build:
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find_advanced/bootstrap find-advanced/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find_by_category/bootstrap find-by-category/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find_by_keywords/bootstrap find-by-keywords/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find_by_product/bootstrap find-by-product/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/find_in_ebay_stores/bootstrap find-in-ebay-stores/main.go

zip: build
	zip -j bin/find_advanced.zip bin/find_advanced/bootstrap
	zip -j bin/find_by_category.zip bin/find_by_category/bootstrap
	zip -j bin/find_by_keywords.zip bin/find_by_keywords/bootstrap
	zip -j bin/find_by_product.zip bin/find_by_product/bootstrap
	zip -j bin/find_in_ebay_stores.zip bin/find_in_ebay_stores/bootstrap

clean:
	rm -rf ./bin

init:
	tofu init

plan: clean zip init
	tofu plan

apply: clean zip init
	tofu apply
