module github.com/etclabscore/ancient-store-s3

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.15
	github.com/ethereum/go-ethereum v1.9.18
	github.com/hashicorp/golang-lru v0.5.4
	github.com/mattn/go-colorable v0.1.7
	github.com/mattn/go-isatty v0.0.12
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/ethereum/go-ethereum => github.com/etclabscore/core-geth v1.11.10-0.20200730130117-dc98713fac98
