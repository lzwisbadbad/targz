package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bcbchain/bclib/tendermint/tmlibs/cli"

	_ "net/http/pprof"

	cmd "github.com/bcbchain/tendermint/cmd/tendermint/commands"
	cfg "github.com/bcbchain/tendermint/config"
	nm "github.com/bcbchain/tendermint/node"
)

func main() {
	go func() {
		if e := http.ListenAndServe(":2020", nil); e != nil {
			fmt.Println("pprof can't start!!!")
		}
	}()

	cmd.AddInitFlags(cmd.InitFilesCmd)
	cmd.AddGenValidatorFlags(cmd.GenValidatorCmd)

	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.InitFilesCmd,
		cmd.ProbeUpnpCmd,
		cmd.ResetAllCmd,
		cmd.ResetPrivValidatorCmd,
		cmd.ShowValidatorCmd,
		cmd.ShowNodeIDCmd,
		cmd.VersionCmd,
		cmd.GenValidatorCmd)

	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc abci app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewNode function
	nodeFunc := nm.DefaultNewNode

	// Create & start node
	rootCmd.AddCommand(
		cmd.NewSyncNodeCmd(nodeFunc),
		cmd.NewRunNodeCmd(nodeFunc))

	command := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
