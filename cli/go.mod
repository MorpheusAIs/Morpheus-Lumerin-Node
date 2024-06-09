module github.com/Lumerin-protocol/Morpheus-Lumerin-Node/cli

go 1.22.0

toolchain go1.22.3

require (
	github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway v0.0.2
	github.com/urfave/cli/v2 v2.27.2
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sashabaranov/go-openai v1.24.1 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	golang.org/x/crypto v0.23.0 // indirect
)

replace github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway => ../api-gateway
