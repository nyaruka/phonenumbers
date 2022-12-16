generated_string.go: phonenumbers.go
	command -v stringer || go install golang.org/x/tools/cmd/stringer
	go generate

build:
	mkdir -p functions
	cd cmd/phoneserver && go build -ldflags "-X main.Version=`git describe --tags`" -o ../../functions/phoneserver .
