package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
	pvm "github.com/bcbchain/tendermint/types/priv_validator"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.
var GenValidatorCmd = &cobra.Command{
	Use:   "gen_validator",
	Short: "Generate new validator keypair",
	Run:   genValidator,
}

func genValidator(cmd *cobra.Command, args []string) {
	chainID, err := cmd.Flags().GetString("chain_id")
	if err != nil {
		fmt.Printf("Generate validator parse chain_id err: %s\n", err)
		return
	}
	if chainID == "" {
		fmt.Printf("Generate validator parse chain_id err: chain_id cannot be empty\n")
		return
	}
	crypto.SetChainId(chainID)

	pv := pvm.GenFilePV(config.PrivValidatorFile())
	pv.Save()
	jsbz, err := cdc.MarshalJSON(pv)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsbz))
}

func AddGenValidatorFlags(cmd *cobra.Command) {
	cmd.Flags().String("chain_id", chainID, "Specify the chain ID")
}
