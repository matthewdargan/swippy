.PHONY: build zip clean deploy

build:
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/ebay-find-by-category/bootstrap ebay-find-by-category/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/ebay-find-by-keyword/bootstrap ebay-find-by-keyword/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/ebay-find-advanced/bootstrap ebay-find-advanced/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/ebay-find-by-product/bootstrap ebay-find-by-product/main.go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/ebay-find-in-ebay-stores/bootstrap ebay-find-in-ebay-stores/main.go

zip:
	zip -j bin/ebay-find-by-category.zip bin/ebay-find-by-category/bootstrap
	zip -j bin/ebay-find-by-keyword.zip bin/ebay-find-by-keyword/bootstrap
	zip -j bin/ebay-find-advanced.zip bin/ebay-find-advanced/bootstrap
	zip -j bin/ebay-find-by-product.zip bin/ebay-find-by-product/bootstrap
	zip -j bin/ebay-find-in-ebay-stores.zip bin/ebay-find-in-ebay-stores/bootstrap

clean:
	rm -rf ./bin

deploy: clean build zip
	sls deploy --verbose
