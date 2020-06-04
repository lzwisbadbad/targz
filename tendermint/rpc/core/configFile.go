package core

import (
	"github.com/bcbchain/bclib/jsoniter"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	core_types "github.com/bcbchain/tendermint/rpc/core/types"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/bcbchain/tendermint/config"
	"github.com/bcbchain/bclib/tendermint/tmlibs/common"
)

var cfg *config.Config = nil
var lock sync.Mutex

func parseConfig() { // 相当不 DRY， 避免循环引用又撸一遍，有空把这个逻辑整合
	if cfg == nil {
		cfg = config.DefaultConfig()
		tmPath := os.Getenv("TMHOME")
		if tmPath == "" {
			home := os.Getenv("HOME")
			if home != "" {
				tmPath = filepath.Join(home, config.DefaultTendermintDir)
			}
		}
		if tmPath == "" {
			tmPath = "/" + config.DefaultTendermintDir
		}
		cfg.SetRoot(tmPath)
	}
}

func GetGenesisPkg(tag string) (*core_types.ResultConfFile, error) {
	if completeStarted == false {
		return nil, errors.New("service not ready")
	}

	lock.Lock()
	defer lock.Unlock()

	if tag == "tmcore" || tag == "" {
		return getGenesisPkgFromTmcore()
	} else if tag == "bcchain" {
		return getGenesisPkgFromBcchain()
	}
	return nil, errors.New("invalid tag, must be tmcore or bcchain")
}

func getGenesisPkgFromTmcore() (*core_types.ResultConfFile, error) {
	parseConfig()

	genesisDir := path.Join(cfg.RootDir, "genesis")
	chainDir := path.Join(genesisDir, genDoc.ChainID)
	targetFile := chainDir + ".tar.gz"

	if !fileutil.Exist(targetFile) {
		err := common.TarIt(chainDir, genesisDir)
		if err != nil {
			return nil, err
		}
		err = common.GzipIt(chainDir+".tar", genesisDir)
		if err != nil {
			return nil, err
		}
	}

	byt, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return nil, err
	}
	jsonBlob, err := jsoniter.Marshal(byt)
	if err != nil {
		return nil, err
	}
	return &core_types.ResultConfFile{F: jsonBlob}, nil
}

func getGenesisPkgFromBcchain() (*core_types.ResultConfFile, error) {
	if completeStarted == false {
		return nil, errors.New("service not ready")
	}

	genesis, err := proxyAppQuery.GetGenesisSync()
	if err != nil {
		return nil, err
	}
	return &core_types.ResultConfFile{F: genesis.Data}, nil
}
