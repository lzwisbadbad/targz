package main

import (
	"flag"
	"os"

	"github.com/bcbchain/bclib/tendermint/go-crypto"
	cmn "github.com/bcbchain/bclib/tendermint/tmlibs/common"
	"github.com/bcbchain/bclib/tendermint/tmlibs/log"

	priv_val "github.com/bcbchain/tendermint/types/priv_validator"
)

func main() {
	var (
		addr        = flag.String("addr", ":46659", "Address of client to connect to")
		chainID     = flag.String("chain-id", "mychain", "chain id")
		privValPath = flag.String("priv", "", "priv val file path")

		logger = log.NewTMLogger("./log", "priv_val_server").With("module", "priv_val")
	)
	flag.Parse()

	logger.Info(
		"Starting private validator",
		"addr", *addr,
		"chainID", *chainID,
		"privPath", *privValPath,
	)

	privVal := priv_val.LoadFilePV(*privValPath)

	rs := priv_val.NewRemoteSigner(
		logger,
		*chainID,
		*addr,
		privVal,
		crypto.GenPrivKeyEd25519(),
	)
	err := rs.Start()
	if err != nil {
		panic(err)
	}

	cmn.TrapSignal(func(sig os.Signal) {
		err := rs.Stop()
		if err != nil {
			panic(err)
		}
	})
}
