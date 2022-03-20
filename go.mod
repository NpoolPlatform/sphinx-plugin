module github.com/NpoolPlatform/sphinx-plugin

go 1.16

require (
	github.com/NpoolPlatform/go-service-framework v0.0.0-20220120091626-4e8035637592
	github.com/NpoolPlatform/message v0.0.0-20220317085427-aeba6e713ae3
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/ethereum/go-ethereum v1.10.16
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-address v0.0.6
	github.com/filecoin-project/go-jsonrpc v0.1.5
	github.com/filecoin-project/go-state-types v0.1.1
	github.com/filecoin-project/lotus v1.13.1
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/ipfs/go-cid v0.1.0
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.7.1-0.20210427113832-6241f9ab9942 // indirect
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.45.0
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.41.0

replace github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.2.6
