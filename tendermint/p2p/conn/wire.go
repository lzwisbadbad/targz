package conn

import (
	"github.com/bcbchain/bclib/tendermint/go-amino"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
	RegisterPacket(cdc)
}
