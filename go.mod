module github.com/kosovopay/go-sdk

go 1.23.0

require (
	github.com/go-resty/resty/v2 v2.17.2
	github.com/google/uuid v1.6.0
)

require golang.org/x/net v0.43.0 // indirect

// v1.0.0 was published under the MIT license. It is superseded by v1.1.0+,
// which are licensed under the KosovoPay License. v1.0.0 is retracted.
retract v1.0.0
