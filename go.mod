module github.com/NpoolPlatform/sphinx-plugin

go 1.16

require (
	github.com/NpoolPlatform/go-service-framework v0.0.0-20211204042807-786e5a616191
	github.com/NpoolPlatform/message v0.0.0-20211205163333-eba28e564885
	github.com/NpoolPlatform/sphinx-proxy v0.0.0-20211206090116-c23ca0d81bb6
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/filecoin-project/go-address v0.0.6
	github.com/filecoin-project/go-jsonrpc v0.1.5
	github.com/filecoin-project/go-state-types v0.1.1
	github.com/filecoin-project/lotus v1.13.1
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/viper v1.9.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.42.0
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.41.0

replace github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.2.6
