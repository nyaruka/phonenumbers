module github.com/nyaruka/phonenumbers/cmd/phoneserver

go 1.19

replace github.com/nyaruka/phonenumbers => ../../

require (
	github.com/aws/aws-lambda-go v1.13.1
	github.com/nyaruka/phonenumbers v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/text v0.3.8 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
