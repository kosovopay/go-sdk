module github.com/kosovopay/go-sdk

go 1.23.0

require (
	github.com/go-resty/resty/v2 v2.17.2
	github.com/google/uuid v1.6.0
)

require golang.org/x/net v0.43.0 // indirect

retract (
	v1.0.0 // Superseded; please use the latest version.
	v1.1.1 // Published under a restrictive license; relicensed to MIT in v1.1.2.
)
