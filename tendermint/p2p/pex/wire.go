package pex

import (
	"github.com/bcbchain/bclib/tendermint/go-amino"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	RegisterPexMessage(cdc)
}
