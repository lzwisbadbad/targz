package p2p

import (
	"github.com/bcbchain/bclib/tendermint/go-amino"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}
