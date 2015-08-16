

all:
	go build

check:
	@go vet
	@golint
