BINARY=argocd-vault-plugin

default: build

quality:
	go vet github.com/argoproj-labs/argocd-vault-plugin
	go test -v -coverprofile cover.out ./...

build:
	go build -o ${BINARY} .

install: build

e2e: install
	./argocd-vault-plugin

getall:
	go run main.go getall -c env.json ./secrets key-quick-start testKey > secrets.yaml
