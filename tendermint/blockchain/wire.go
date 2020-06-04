package blockchain

import (
	"github.com/bcbchain/bclib/tendermint/go-amino"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	RegisterBlockchainMessages(cdc)
	crypto.RegisterAmino(cdc)
}
