package mempool

import (
	"github.com/bcbchain/bclib/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterMempoolMessages(cdc)
}
