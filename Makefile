help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test:
	go test -v ./...

build:
	env GOOS=linux go build -ldflags="-s -w" -o ./bom-observations-exporter

deploy-staging:
	sls deploy --stage=dev --verbose

deploy-prod:
	sls deploy --stage=prod --verbose