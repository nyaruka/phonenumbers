module github.com/nyaruka/phonenumbers/cmd/phoneserver

go 1.12

replace github.com/nyaruka/phonenumbers => ../../

require (
	github.com/aws/aws-lambda-go v1.13.1
	github.com/nyaruka/phonenumbers v0.0.0-00010101000000-000000000000
)
