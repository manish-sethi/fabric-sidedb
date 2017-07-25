/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package privacyenabledstate

import (
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
)

// DBProvider provides handle to a PvtVersionedDB
type DBProvider interface {
	// GetDBHandle returns a handle to a PvtVersionedDB
	GetDBHandle(id string) (DB, error)
	// Close closes all the PvtVersionedDB instances and releases any resources held by VersionedDBProvider
	Close()
}

// DB extends VersionedDB interface. This interface provides additional functions for managing private data state
type DB interface {
	statedb.VersionedDB
	GetPrivateData(namespace, collection, key string) (*statedb.VersionedValue, error)
	GetValueHash(namespace, collection string, keyHash []byte) (*statedb.VersionedValue, error)
	GetPrivateDataMultipleKeys(namespace, collection string, keys []string) ([]*statedb.VersionedValue, error)
	GetPrivateDataRangeScanIterator(namespace, collection, startKey, endKey string) (statedb.ResultsIterator, error)
	ExecuteQueryOnPrivateData(namespace, collection, query string) (statedb.ResultsIterator, error)
	ApplyPrivacyAwareUpdates(updates *UpdateBatch, height *version.Height) error
}

// UpdateBatch encapsulates the updates to Public, Private, and Hashed data.
// This is expected to contain a consistent set of updates
type UpdateBatch struct {
	PubUpdates  *PubUpdateBatch
	HashUpdates *HashedUpdateBatch
	PvtUpdates  *PvtUpdateBatch
}

// PubUpdateBatch contains update for the public data
type PubUpdateBatch struct {
	*statedb.UpdateBatch
}

// HashedUpdateBatch contains updates for the hashes of the private data
type HashedUpdateBatch struct {
	UpdateMap
}

// PvtUpdateBatch contains updates for the private data
type PvtUpdateBatch struct {
	UpdateMap
}

// UpdateMap maintains entries of tuple <Namespace, UpdatesForNamespace>
type UpdateMap map[string]nsBatch

// nsBatch contains updates related to one namespace
type nsBatch struct {
	*statedb.UpdateBatch
}

// NewUpdateBatch creates and empty UpdateBatch
func NewUpdateBatch() *UpdateBatch {
	return &UpdateBatch{NewPubUpdateBatch(), NewHashedUpdateBatch(), NewPvtUpdateBatch()}
}

// NewPubUpdateBatch creates an empty PubUpdateBatch
func NewPubUpdateBatch() *PubUpdateBatch {
	return &PubUpdateBatch{statedb.NewUpdateBatch()}
}

// NewHashedUpdateBatch creates an empty HashedUpdateBatch
func NewHashedUpdateBatch() *HashedUpdateBatch {
	return &HashedUpdateBatch{make(map[string]nsBatch)}
}

// NewPvtUpdateBatch creates an empty PvtUpdateBatch
func NewPvtUpdateBatch() *PvtUpdateBatch {
	return &PvtUpdateBatch{make(map[string]nsBatch)}
}

// IsEmpty returns true if there exists any updates
func (b UpdateMap) IsEmpty() bool {
	return len(b) == 0
}

// Put sets the value in the batch for a given combination of namespace and collection name
func (b UpdateMap) Put(ns, coll, key string, value []byte, version *version.Height) {
	b.getOrCreateNsBatch(ns).Put(coll, key, value, version)
}

// Delete removes the entry from the batch for a given combination of namespace and collection name
func (b UpdateMap) Delete(ns, coll, key string, version *version.Height) {
	b.getOrCreateNsBatch(ns).Delete(coll, key, version)
}

// Get retrieves the value from the bacth for a given combination of namespace and collection name
func (b UpdateMap) Get(ns, coll, key string) *statedb.VersionedValue {
	nsPvtBatch, ok := b[ns]
	if !ok {
		return nil
	}
	return nsPvtBatch.Get(coll, key)
}

func (nsb nsBatch) GetCollectionNames() []string {
	return nsb.GetUpdatedNamespaces()
}

func (b UpdateMap) getOrCreateNsBatch(ns string) nsBatch {
	batch, ok := b[ns]
	if !ok {
		batch = nsBatch{statedb.NewUpdateBatch()}
		b[ns] = batch
	}
	return batch
}

// Contains returns true if the given <ns,coll,keyHash> tuple is present in the batch
func (h HashedUpdateBatch) Contains(ns, coll string, keyHash []byte) bool {
	nsBatch, ok := h.UpdateMap[ns]
	if !ok {
		return false
	}
	return nsBatch.Exists(coll, string(keyHash))
}

// Put overrides the function in UpdateMap for allowing the key to be a []byte instead of a string
func (h HashedUpdateBatch) Put(ns, coll string, key []byte, value []byte, version *version.Height) {
	h.UpdateMap.Put(ns, coll, string(key), value, version)
}

// Delete overrides the function in UpdateMap for allowing the key to be a []byte instead of a string
func (h HashedUpdateBatch) Delete(ns, coll string, key []byte, version *version.Height) {
	h.UpdateMap.Delete(ns, coll, string(key), version)
}
