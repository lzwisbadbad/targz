package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
	"github.com/bcbchain/tendermint/p2p"
	"github.com/bcbchain/tendermint/types"
)

// ShowNodeIDCmd dumps node's ID to the standard output.
var ShowNodeIDCmd = &cobra.Command{
	Use:   "show_node_id",
	Short: "Show this node's ID",
	RunE:  showNodeID,
}

func showNodeID(cmd *cobra.Command, args []string) error {

	genDoc, err := types.GenesisDocFromFile(config)
	if err != nil {
		logger.Error("tendermint can't parse genesis file", "parse", err)
		return err
	}
	crypto.SetChainId(genDoc.ChainID)

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())

	return nil
}
