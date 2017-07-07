/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pvtrwstorage

import (
	"fmt"
	"os"
	"testing"

	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	viper.Set("peer.fileSystemPath", "/tmp/fabric/ledgertests/kvledger/")
	os.Exit(m.Run())
}

func TestPurgeIndexKeyCodingEncoding(t *testing.T) {
	assert := assert.New(t)
	blkHts := []uint64{0, 10, 20000}
	txids := []string{"txid", ""}
	endorserids := []string{"endorserid", ""}
	for _, blkHt := range blkHts {
		for _, txid := range txids {
			for _, endorserid := range endorserids {
				testCase := fmt.Sprintf("blkHt=%d,txid=%s,endorserid=%s", blkHt, txid, endorserid)
				t.Run(testCase, func(t *testing.T) {
					t.Logf("Running test case [%s]", testCase)
					purgeIndexKey := createCompositeKeyForPurgeIndex(blkHt, txid, endorserid)
					txid1, endorserid1, blkHt1 := splitCompositeKeyOfPurgeIndex(purgeIndexKey)
					assert.Equal(txid, txid1)
					assert.Equal(endorserid, endorserid1)
					assert.Equal(blkHt, blkHt1)
				})
			}
		}
	}
}

func TestRWSetKeyCodingEncoding(t *testing.T) {
	assert := assert.New(t)
	blkHts := []uint64{0, 10, 20000}
	txids := []string{"txid", ""}
	endorserids := []string{"endorserid", ""}
	for _, blkHt := range blkHts {
		for _, txid := range txids {
			for _, endorserid := range endorserids {
				testCase := fmt.Sprintf("blkHt=%d,txid=%s,endorserid=%s", blkHt, txid, endorserid)
				t.Run(testCase, func(t *testing.T) {
					t.Logf("Running test case [%s]", testCase)
					rwsetKey := createCompositeKeyForPRWSet(txid, endorserid, blkHt)
					endorserid1, blkHt1 := splitCompositeKeyOfPRWSet(rwsetKey)
					assert.Equal(endorserid, endorserid1)
					assert.Equal(blkHt, blkHt1)
				})
			}
		}
	}
}

func TestTransientStorePersistAndRetrieve(t *testing.T) {
	env := newTestTransientStoreEnv(t)
	assert := assert.New(t)
	txid := "txid-1"

	// Create private simulation results for txid-1
	var endorsersResults []*ledger.EndorserPrivateSimulationResults

	// Results produced by endorser 1
	endorser0SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser0",
		EndorsementBlockHeight:   10,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser0SimulationResults)

	// Results produced by endorser 2
	endorser1SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser1",
		EndorsementBlockHeight:   10,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser1SimulationResults)

	// Persist simulation results into transient store
	var err error
	for i := 0; i < len(endorsersResults); i++ {
		err = env.testTransientStore.Persist(txid, endorsersResults[i].EndorserId,
			endorsersResults[i].EndorsementBlockHeight, endorsersResults[i].PrivateSimulationResults)
		assert.NoError(err)
	}

	// Retrieve simulation results of txid-1 from transient store
	var iter commonledger.ResultsIterator
	iter, err = env.testTransientStore.GetTxPrivateRWSetByTxid(txid)
	assert.NoError(err)

	var result commonledger.QueryResult
	var actualEndorsersResults []*ledger.EndorserPrivateSimulationResults
	for true {
		result, err = iter.Next()
		assert.NoError(err)
		if result == nil {
			break
		}
		actualEndorsersResults = append(actualEndorsersResults, result.(*ledger.EndorserPrivateSimulationResults))
	}
	iter.Close()
	assert.Equal(endorsersResults, actualEndorsersResults)
}

func TestTransientStorePurge(t *testing.T) {
	env := newTestTransientStoreEnv(t)
	assert := assert.New(t)

	txid := "txid-1"

	// Create private simulation results for txid-1
	var endorsersResults []*ledger.EndorserPrivateSimulationResults

	// Results produced by endorser 1
	endorser0SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser0",
		EndorsementBlockHeight:   10,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser0SimulationResults)

	// Results produced by endorser 2
	endorser1SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser1",
		EndorsementBlockHeight:   11,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser1SimulationResults)

	// Results produced by endorser 3
	endorser2SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser2",
		EndorsementBlockHeight:   12,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser2SimulationResults)

	// Results produced by endorser 3
	endorser3SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser3",
		EndorsementBlockHeight:   12,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser3SimulationResults)

	// Results produced by endorser 3
	endorser4SimulationResults := &ledger.EndorserPrivateSimulationResults{
		EndorserId:               "endorser4",
		EndorsementBlockHeight:   13,
		PrivateSimulationResults: []byte("results"),
	}
	endorsersResults = append(endorsersResults, endorser4SimulationResults)

	// Persist simulation results into transient store
	var err error
	for i := 0; i < 5; i++ {
		err = env.testTransientStore.Persist(txid, endorsersResults[i].EndorserId,
			endorsersResults[i].EndorsementBlockHeight, endorsersResults[i].PrivateSimulationResults)
		assert.NoError(err)
	}

	// Retain results generate at block height greater than or equal to 12
	minEndorsementBlkHtToRetain := uint64(12)
	err = env.testTransientStore.Purge(minEndorsementBlkHtToRetain)
	assert.NoError(err)

	// Retrieve simulation results of txid-1 from transient store
	var iter commonledger.ResultsIterator
	iter, err = env.testTransientStore.GetTxPrivateRWSetByTxid(txid)
	assert.NoError(err)

	// Expected results for txid-1
	var expectedEndorsersResults []*ledger.EndorserPrivateSimulationResults
	expectedEndorsersResults = append(expectedEndorsersResults, endorser2SimulationResults) //endorsed at height 12
	expectedEndorsersResults = append(expectedEndorsersResults, endorser3SimulationResults) //endorsed at height 12
	expectedEndorsersResults = append(expectedEndorsersResults, endorser4SimulationResults) //endorsed at height 13

	// Check whether actual results and expected results are same
	var result commonledger.QueryResult
	var actualEndorsersResults []*ledger.EndorserPrivateSimulationResults
	for true {
		result, err = iter.Next()
		assert.NoError(err)
		if result == nil {
			break
		}
		actualEndorsersResults = append(actualEndorsersResults, result.(*ledger.EndorserPrivateSimulationResults))
	}
	iter.Close()
	assert.Equal(expectedEndorsersResults, actualEndorsersResults)

	// Get the minimum retained endorsement block height
	var actualMinEndorsementBlkHt uint64
	actualMinEndorsementBlkHt, err = env.testTransientStore.GetMinEndorsementBlkHt()
	assert.NoError(err)
	assert.Equal(minEndorsementBlkHtToRetain, actualMinEndorsementBlkHt)

	// Retain results generate at block height greater than or equal to 15
	minEndorsementBlkHtToRetain = uint64(15)
	err = env.testTransientStore.Purge(minEndorsementBlkHtToRetain)
	assert.NoError(err)

	// There should be no entries in the transient store
	actualMinEndorsementBlkHt, err = env.testTransientStore.GetMinEndorsementBlkHt()
	assert.Equal(err, ErrTransientStoreEmpty)

	// Retain results generate at block height greater than or equal to 15
	minEndorsementBlkHtToRetain = uint64(15)
	err = env.testTransientStore.Purge(minEndorsementBlkHtToRetain)
	// Should not return any error
	assert.NoError(err)

	env.cleanup()
}
