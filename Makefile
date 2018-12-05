build:
	mkdir -p functions
	go get ./...
	go build -o functions/phoneserver ./cmd/phoneserver
