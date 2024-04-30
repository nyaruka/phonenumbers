module github.com/lytics/phonenumbers/cmd/phoneserver

go 1.19

replace github.com/lytics/phonenumbers => ../../

require (
	github.com/aws/aws-lambda-go v1.13.1
	github.com/lytics/phonenumbers v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
