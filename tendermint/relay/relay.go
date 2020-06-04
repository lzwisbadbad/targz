package relay

import (
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	types2 "github.com/bcbchain/bclib/tendermint/abci/types"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
	cfg "github.com/bcbchain/tendermint/config"
	"github.com/bcbchain/tendermint/proxy"
	"github.com/bcbchain/tendermint/types"
	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
	"strings"
	"sync"
)

type RelayController struct {
	LocalURL         string   // local url
	ChainIDToURLs    sync.Map // chainID => openURLs
	ChainIDToAddrVer sync.Map // chainID => addressVersion

	QueueIDToQueueRelay map[string]*QueueRelay // queueID => QueueRelay

	currentNodeAddress string
	config             *cfg.Config
	abciClient         proxy.AppConns
	logger             log.Logger
}

var (
	gRelay   *RelayController
	initOnce sync.Once
)

// Init init relay controller
func Init(config *cfg.Config, logger log.Logger, conns proxy.AppConns) *RelayController {
	initOnce.Do(func() {
		temp := strings.Split(config.RPC.ListenAddress, ":")
		localURL := "http://127.0.0.1:" + temp[len(temp)-1]

		gRelay = &RelayController{
			LocalURL:           localURL,
			currentNodeAddress: getNodeAddress(config, "", "", 0),
			config:             config,
			abciClient:         conns,
			logger:             logger,
		}

		gRelay.init()

		logger.Info("RELAY init", "gRelay", gRelay)
	})

	return gRelay
}

// GetRelayController get instance
func GetRelayController() *RelayController {
	return gRelay
}

// SetNewHeader determines whether to start or stop a relay by header
func (rc *RelayController) SetNewHeader(header *types.Header) {
	if header.Relayer == nil {
		return
	}

	if header.Relayer.Address == rc.currentNodeAddress {
		rc.logger.Debug("RELAY SetNewHeader", "rc.startRelay()")
		rc.startRelay()
	} else {
		rc.logger.Debug("RELAY SetNewHeader", "rc.stopRelay()",
			fmt.Sprintf("expcted: %s obtain: %s", header.Relayer.Address, rc.currentNodeAddress))
		rc.stopRelay()
	}
}

// UpdateOpenURL update relay controller.ChainIDToURLS, overwrite existing data.
func (rc *RelayController) UpdateOpenURL(chainID string, urls []string, addrVer int32) {
	rc.logger.Info("RELAY UpdateOpenURL", "chainID", chainID, "urls", urls)
	localChainID := rc.getLocalChainID()

	if localChainID == chainID {
		return
	}

	if _, ok := rc.ChainIDToURLs.Load(chainID); ok {
		queueID := makeQueueID(localChainID, chainID)
		qr, ok := rc.QueueIDToQueueRelay[queueID]
		if !ok {
			return
		}
		qr.RemoteURLs = urls
		rc.QueueIDToQueueRelay[queueID] = qr

	} else {
		rc.addQueueRelay(localChainID, chainID, urls, addrVer)
	}

	rc.ChainIDToURLs.Store(chainID, urls)
}

// relayControler initialize
func (rc *RelayController) init() {
	localChainID := rc.getLocalChainID()
	if len(localChainID) == 0 {
		panic("can not get local chainID.")
	}

	if strings.Contains(localChainID, "[") {
		// side chain
		rc.getMainChainURLs(getMainChaidID(localChainID))
	} else {
		// main chain
		rc.getSideChainOpenURL()
	}
}

func (rc *RelayController) addQueueRelay(localChainID, toChainID string, urls []string, addrVer int32) {
	queueID := makeQueueID(localChainID, toChainID)
	qr := QueueRelay{
		LocalURL:     rc.LocalURL,
		RemoteURLs:   urls,
		QueueID:      queueID,
		genesisOrgID: gRelay.queryGenesisOrgID(),
		signalChan:   make(chan bool, 100),
		currentNode:  rc.getCurrentNode(queueID, addrVer),
		logger:       rc.logger,
	}

	go qr.Start()
	qr.signalChan <- true

	if len(rc.QueueIDToQueueRelay) == 0 {
		rc.QueueIDToQueueRelay = make(map[string]*QueueRelay)
	}
	rc.QueueIDToQueueRelay[qr.QueueID] = &qr

	rc.logger.Debug("RELAY addQueueRelay", "queueRelay", qr)
}

func (rc *RelayController) getMainChainURLs(mainChainID string) {
	var urls []string
	rc.abciQueryAndParse(keyOfOpenURLs(mainChainID), &urls)
	if len(urls) == 0 {
		panic("can not get main chain URL")
	}

	rc.ChainIDToURLs.Store(mainChainID, urls)
}

func (rc *RelayController) getSideChainOpenURL() {
	sideChainIDs := rc.getSideChainIDs()

	for _, chainID := range sideChainIDs {

		status, addrVer := rc.getSideChainStatusAndAddrVer(chainID)
		if status != "ready" && status != "clear" {
			continue
		}

		urls := rc.getOepnURLs(chainID)

		rc.ChainIDToURLs.Store(chainID, urls)
		rc.ChainIDToAddrVer.Store(chainID, addrVer)
	}
}

func (rc *RelayController) getOepnURLs(chainID string) []string {
	urls := new([]string)
	rc.abciQueryAndParse(keyOfOpenURLs(chainID), urls)
	return *urls
}

func (rc *RelayController) getSideChainIDs() []string {
	sideChainIDs := new([]string)
	rc.abciQueryAndParse(keyOfSideChainIDs(), &sideChainIDs)
	return *sideChainIDs
}

func (rc *RelayController) getSideChainStatusAndAddrVer(chainID string) (string, AddressVersion) {
	ci := new(ChainInfo)
	rc.abciQueryAndParse(keyOfChainInfo(chainID), ci)
	return ci.Status, ci.AddrVersion
}

func (rc *RelayController) abciQueryAndParse(key string, data interface{}) {
	r, err := rc.abciClient.Query().QuerySync(types2.RequestQuery{
		Path: key,
	})
	if err != nil {
		panic(err)
	}

	if len(r.GetValue()) == 0 {
		return
	}

	_ = jsoniter.Unmarshal(r.GetValue(), data)
}

func (rc *RelayController) startRelay() {
	localChainID := rc.getLocalChainID()

	if len(rc.QueueIDToQueueRelay) == 0 {
		rc.QueueIDToQueueRelay = make(map[string]*QueueRelay)
		rc.ChainIDToURLs.Range(func(chanID, urls interface{}) bool {
			queueID := makeQueueID(localChainID, chanID.(string))

			var addrVer AddressVersion
			v, ok := rc.ChainIDToAddrVer.Load(chanID)
			if ok {
				addrVer = v.(AddressVersion)
			}

			qr := QueueRelay{
				LocalURL:     rc.LocalURL,
				RemoteURLs:   urls.([]string),
				QueueID:      queueID,
				genesisOrgID: gRelay.queryGenesisOrgID(),
				signalChan:   make(chan bool, 100),
				currentNode:  rc.getCurrentNode(queueID, int32(addrVer)),
				logger:       rc.logger,
			}
			rc.QueueIDToQueueRelay[qr.QueueID] = &qr

			go qr.Start()
			qr.signalChan <- true
			return true
		})

	} else {
		for _, v := range rc.QueueIDToQueueRelay {
			v.signalChan <- true
		}
	}
}

func (rc *RelayController) stopRelay() {
	for _, v := range rc.QueueIDToQueueRelay {
		v.signalChan <- false
	}
}

func (rc *RelayController) getLocalChainID() string {
	chainID := new(string)
	r, e := rc.abciClient.Query().QuerySync(types2.RequestQuery{
		Path: keyOfChainID(),
	})
	if e != nil {
		rc.logger.Error("RELAY", "can not get local chainID", e)
		return ""
	}

	e = jsoniter.Unmarshal(r.GetValue(), chainID)
	if e != nil {
		// 正式链 1.0 和 2.0 的格式不一样
		return string(r.GetValue())
	}
	return *chainID
}

func (rc *RelayController) getCurrentNode(queueID string, addrVer int32) *CurrentNodeInfo {
	privKey := getCurrentNodePrivKey(rc.config)
	priKey := privKey.(crypto.PrivKeyEd25519)
	p := "0x" + hex.EncodeToString(priKey[:])

	fromChainID, toChainID := splitQueueID(queueID)

	currentNodeInfo := &CurrentNodeInfo{
		Address:    getNodeAddress(rc.config, fromChainID, toChainID, addrVer),
		HexPrivKey: p,
		Nonce:      0,
	}
	return currentNodeInfo
}

func (rc *RelayController) queryGenesisOrgID() string {
	r, e := rc.abciClient.Query().QuerySync(types2.RequestQuery{
		Path: keyOfGenesisOrgID(),
	})
	if e != nil {
		rc.logger.Error("RELAY", "can not get local genesis org ID", e)
		return ""
	}

	genesisOrgID := new(string)
	e = jsoniter.Unmarshal(r.GetValue(), genesisOrgID)
	if e != nil {
		return ""
	}
	return *genesisOrgID
}
