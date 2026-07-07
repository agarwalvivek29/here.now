module github.com/agarwalvivek29/here.now/services/herenow-api

go 1.26.3

require (
	github.com/agarwalvivek29/here.now/packages/schema/generated/go v0.0.0
	google.golang.org/protobuf v1.36.11
)

replace github.com/agarwalvivek29/here.now/packages/schema/generated/go => ../../packages/schema/generated/go
