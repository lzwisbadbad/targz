package state

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bcbchain/tendermint/proxy"
	"github.com/bcbchain/tendermint/types"
	dbm "github.com/bcbchain/bclib/tendermint/tmlibs/db"
)

//-----------------------------------------------------
// Validate block

func validateBlock(proxyApp proxy.AppConnConsensus, stateDB, stateDBx dbm.DB, s State, b *types.Block) error {
	// if block version great than local version, return error for exit application wait for upgrade
	//if b.BlockVersion != nil && versions.GreaterThan(*b.BlockVersion, types.BlockVersion) {
	//	// 只判断 block version。程序版本不判断。
	//	return fmt.Errorf("local block version too old , expected %v, got %v", *b.BlockVersion, types.BlockVersion)
	//}

	// validate internal consistency
	if err := b.ValidateBasic(); err != nil {
		return err
	}

	// validate basic info
	if b.ChainID != s.ChainID {
		return fmt.Errorf("Wrong Block.Header.ChainID. Expected %v, got %v", s.ChainID, b.ChainID)
	}
	if b.Height != s.LastBlockHeight+1 {
		return fmt.Errorf("Wrong Block.Header.Height. Expected %v, got %v", s.LastBlockHeight+1, b.Height)
	}
	/*	TODO: Determine bounds for Time
		See blockchain/reactor "stopSyncingDurationMinutes"

		if !b.Time.After(lastBlockTime) {
			return errors.New("Invalid Block.Header.Time")
		}
	*/

	// validate prev block info
	if !b.LastBlockID.Equals(s.LastBlockID) {
		return fmt.Errorf("Wrong Block.Header.LastBlockID.  Expected %v, hieght:%d, got %v, height:%d", s.LastBlockID, s.LastBlockHeight, b.LastBlockID, b.Height)
	}
	newTxs := int64(len(b.Data.Txs))
	if b.TotalTxs != s.LastBlockTotalTx+newTxs {
		return fmt.Errorf("Wrong Block.Header.TotalTxs. Expected %v, got %v", s.LastBlockTotalTx+newTxs, b.TotalTxs)
	}

	// validate app info
	if !bytes.Equal(b.LastAppHash, s.LastAppHash) {
		// rollback bcchain & tendermint one block height
		rollback(proxyApp, stateDBx)
		return fmt.Errorf("Wrong Block.Header.AppHash.  Expected %X, got %v", s.LastAppHash, b.LastAppHash)
	}
	if !bytes.Equal(b.ConsensusHash, s.ConsensusParams.Hash()) {
		return fmt.Errorf("Wrong Block.Header.ConsensusHash.  Expected %X, got %v", s.ConsensusParams.Hash(), b.ConsensusHash)
	}
	if !bytes.Equal(b.LastResultsHash, s.LastResultsHash) {
		return fmt.Errorf("Wrong Block.Header.LastResultsHash.  Expected %X, got %v", s.LastResultsHash, b.LastResultsHash)
	}
	if !bytes.Equal(b.ValidatorsHash, s.Validators.Hash()) {
		return fmt.Errorf("Wrong Block.Header.ValidatorsHash.  Expected %X, got %v", s.Validators.Hash(), b.ValidatorsHash)
	}

	// Validate block LastCommit.
	if b.Height == 1 {
		if len(b.LastCommit.Precommits) != 0 {
			return errors.New("Block at height 1 (first block) should have no LastCommit precommits")
		}
	} else {
		if len(b.LastCommit.Precommits) != s.LastValidators.Size() {
			return fmt.Errorf("Invalid block commit size. Expected %v, got %v",
				s.LastValidators.Size(), len(b.LastCommit.Precommits))
		}
		err := s.LastValidators.VerifyCommit(
			s.ChainID, s.LastBlockID, b.Height-1, b.LastCommit)
		if err != nil {
			return err
		}
	}

	for _, ev := range b.Evidence.Evidence {
		if err := VerifyEvidence(stateDB, s, ev); err != nil {
			return types.NewEvidenceInvalidErr(ev, err)
		}
	}

	return nil
}

// XXX: What's cheaper (ie. what should be checked first):
//  evidence internal validity (ie. sig checks) or validator existed (fetch historical val set from db)

// VerifyEvidence verifies the evidence fully by checking it is internally
// consistent and sufficiently recent.
func VerifyEvidence(stateDB dbm.DB, s State, evidence types.Evidence) error {
	height := s.LastBlockHeight

	evidenceAge := height - evidence.Height()
	maxAge := s.ConsensusParams.EvidenceParams.MaxAge
	if evidenceAge > maxAge {
		return fmt.Errorf("Evidence from height %d is too old. Min height is %d",
			evidence.Height(), height-maxAge)
	}

	if err := evidence.Verify(s.ChainID); err != nil {
		return err
	}

	valset, err := LoadValidators(stateDB, evidence.Height())
	if err != nil {
		// TODO: if err is just that we cant find it cuz we pruned, ignore.
		// TODO: if its actually bad evidence, punish peer
		return err
	}

	// The address must have been an active validator at the height
	ev := evidence
	height, addr, idx := ev.Height(), ev.Address(), ev.Index()
	valIdx, val := valset.GetByAddress(addr)
	if val == nil {
		return fmt.Errorf("Address %X was not a validator at height %d", addr, height)
	} else if idx != valIdx {
		return fmt.Errorf("Address %X was validator %d at height %d, not %d", addr, valIdx, height, idx)
	}

	return nil
}

func rollback(proxyApp proxy.AppConnConsensus, stateDBx dbm.DB) {
	// rollback bcchain
	r, e := proxyApp.RollbackSync()
	if e != nil {
		panic(e)
	}
	if r.Code != 200 {
		panic(r.Log)
	}

	// rollback state to last state
	s := LoadLastState(stateDBx)
	SaveState(stateDBx, s)

	// storeBlockHeight - 1
	storeState := LoadBlockStoreStateJSON(stateDBx)
	storeState.Height = storeState.Height - 1
	storeState.Save(stateDBx)
}

// 此处和 blockchain 包中一致，为了避免循环引用，所以这里也写了，以后有修改需要两处同步。
var blockStoreKey = []byte("blockStore")

type BlockStoreStateJSON struct {
	Height int64 `json:"height"`
}

// Save persists the blockStore state to the database as JSON.
func (bsj BlockStoreStateJSON) Save(db dbm.DB) {
	byt, err := cdc.MarshalJSON(bsj)
	if err != nil {
		panic("Could not marshal state bytes:" + err.Error())
	}
	db.SetSync(blockStoreKey, byt)
}

// LoadBlockStoreStateJSON returns the BlockStoreStateJSON as loaded from disk.
// If no BlockStoreStateJSON was previously persisted, it returns the zero value.
func LoadBlockStoreStateJSON(db dbm.DB) BlockStoreStateJSON {
	byt := db.Get(blockStoreKey)
	if len(byt) == 0 {
		return BlockStoreStateJSON{
			Height: 0,
		}
	}
	bsj := BlockStoreStateJSON{}
	err := cdc.UnmarshalJSON(byt, &bsj)
	if err != nil {
		panic(fmt.Sprintf("Could not unmarshal bytes: %X", byt))
	}
	return bsj
}
