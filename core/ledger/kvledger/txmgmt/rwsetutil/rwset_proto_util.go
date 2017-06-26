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

package rwsetutil

import (
	"github.com/golang/protobuf/proto"
	bccspfactory "github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
)

/////////////////////////////////////////////////////////////////
// Messages related to PUBLIC read-write set
/////////////////////////////////////////////////////////////////

// TxRwSet acts as a proxy of 'rwset.TxReadWriteSet' proto message and helps constructing Read-write set specifically for KV data model
type TxRwSet struct {
	NsRwSets []*NsRwSet
}

// NsRwSet encapsulates 'kvrwset.KVRWSet' proto message for a specific name space (chaincode)
type NsRwSet struct {
	NameSpace       string
	KvRwSet         *kvrwset.KVRWSet
	CollHashedRwSet []*CollHashedRwSet
}

// CollHashedRwSet encapsulates 'kvrwset.HashedRWSet' proto message for a specific collection
type CollHashedRwSet struct {
	CollectionName string
	HashedRwSet    *kvrwset.HashedRWSet
	PvtRwSetHash   []byte
}

/////////////////////////////////////////////////////////////////
// Messages related to PRIVATE read-write set
/////////////////////////////////////////////////////////////////

// TxPvtRwSet represents 'rwset.TxPvtReadWriteSet' proto message
type TxPvtRwSet struct {
	NsPvtRwSet []*NsPvtRwSet
}

// NsPvtRwSet represents 'rwset.NsPvtReadWriteSet' proto message
type NsPvtRwSet struct {
	NameSpace    string
	CollPvtRwSet []*CollPvtRwSet
}

// CollPvtRwSet encapsulates 'kvrwset.KVRWSet' proto message for a private rwset for a specific collection
// KvRwSet in a private RwSet should not contain range query info
type CollPvtRwSet struct {
	CollectionName string
	KvRwSet        *kvrwset.KVRWSet
}

/////////////////////////////////////////////////////////////////
// FUNCTIONS for converting messages to/from proto bytes
/////////////////////////////////////////////////////////////////

// ToProtoBytes constructs TxReadWriteSet proto message and serializes using protobuf Marshal
func (txRwSet *TxRwSet) ToProtoBytes() ([]byte, error) {
	var protoMsg *rwset.TxReadWriteSet
	var err error
	if protoMsg, err = txRwSet.toProtoMsg(); err != nil {
		return nil, err
	}
	return proto.Marshal(protoMsg)
}

// FromProtoBytes deserializes protobytes into TxReadWriteSet proto message and populates 'TxRwSet'
func (txRwSet *TxRwSet) FromProtoBytes(protoBytes []byte) error {
	protoMsg := &rwset.TxReadWriteSet{}
	var err error
	var txRwSetTemp *TxRwSet
	if err = proto.Unmarshal(protoBytes, protoMsg); err != nil {
		return err
	}
	if txRwSetTemp, err = txRwSetFromProtoMsg(protoMsg); err != nil {
		return err
	}
	txRwSet.NsRwSets = txRwSetTemp.NsRwSets
	return nil
}

func (txRwSet *TxRwSet) toProtoMsg() (*rwset.TxReadWriteSet, error) {
	protoMsg := &rwset.TxReadWriteSet{DataModel: rwset.TxReadWriteSet_KV}
	var nsRwSetProtoMsg *rwset.NsReadWriteSet
	var err error
	for _, nsRwSet := range txRwSet.NsRwSets {
		if nsRwSetProtoMsg, err = nsRwSet.toProtoMsg(); err != nil {
			return nil, err
		}
		protoMsg.NsRwset = append(protoMsg.NsRwset, nsRwSetProtoMsg)
	}
	return protoMsg, nil
}

func txRwSetFromProtoMsg(protoMsg *rwset.TxReadWriteSet) (*TxRwSet, error) {
	txRwSet := &TxRwSet{}
	var nsRwSet *NsRwSet
	var err error
	for _, nsRwSetProtoMsg := range protoMsg.NsRwset {
		if nsRwSet, err = nsRwSetFromProtoMsg(nsRwSetProtoMsg); err != nil {
			return nil, err
		}
		txRwSet.NsRwSets = append(txRwSet.NsRwSets, nsRwSet)
	}
	return txRwSet, nil
}

func (nsRwSet *NsRwSet) toProtoMsg() (*rwset.NsReadWriteSet, error) {
	var err error
	protoMsg := &rwset.NsReadWriteSet{Namespace: nsRwSet.NameSpace}
	if protoMsg.Rwset, err = proto.Marshal(nsRwSet.KvRwSet); err != nil {
		return nil, err
	}

	var collHashedRwSetProtoMsg *rwset.CollectionHashedReadWriteSet
	for _, collHashedRwSet := range nsRwSet.CollHashedRwSet {
		if collHashedRwSetProtoMsg, err = collHashedRwSet.toProtoMsg(); err != nil {
			return nil, err
		}
		protoMsg.CollectionHashedRwset = append(protoMsg.CollectionHashedRwset, collHashedRwSetProtoMsg)
	}
	return protoMsg, nil
}

func nsRwSetFromProtoMsg(protoMsg *rwset.NsReadWriteSet) (*NsRwSet, error) {
	nsRwSet := &NsRwSet{NameSpace: protoMsg.Namespace, KvRwSet: &kvrwset.KVRWSet{}}
	if err := proto.Unmarshal(protoMsg.Rwset, nsRwSet.KvRwSet); err != nil {
		return nil, err
	}
	var err error
	var collHashedRwSet *CollHashedRwSet
	for _, collHashedRwSetProtoMsg := range protoMsg.CollectionHashedRwset {
		if collHashedRwSet, err = collHashedRwSetFromProtoMsg(collHashedRwSetProtoMsg); err != nil {
			return nil, err
		}
		nsRwSet.CollHashedRwSet = append(nsRwSet.CollHashedRwSet, collHashedRwSet)
	}
	return nsRwSet, nil
}

func (collHashedRwSet *CollHashedRwSet) toProtoMsg() (*rwset.CollectionHashedReadWriteSet, error) {
	var err error
	protoMsg := &rwset.CollectionHashedReadWriteSet{
		CollectionName: collHashedRwSet.CollectionName,
		PvtRwsetHash:   collHashedRwSet.PvtRwSetHash,
	}
	if protoMsg.HashedRwset, err = proto.Marshal(collHashedRwSet.HashedRwSet); err != nil {
		return nil, err
	}
	return protoMsg, nil
}

func collHashedRwSetFromProtoMsg(protoMsg *rwset.CollectionHashedReadWriteSet) (*CollHashedRwSet, error) {
	colHashedRwSet := &CollHashedRwSet{
		CollectionName: protoMsg.CollectionName,
		PvtRwSetHash:   protoMsg.PvtRwsetHash,
		HashedRwSet:    &kvrwset.HashedRWSet{},
	}
	if err := proto.Unmarshal(protoMsg.HashedRwset, colHashedRwSet.HashedRwSet); err != nil {
		return nil, err
	}
	return colHashedRwSet, nil
}

///////////////////////////////////////////////////////////////////////////////
// functions for private read-write set
///////////////////////////////////////////////////////////////////////////////

func (txPvtRwSet *TxPvtRwSet) toProtoMsg() (*rwset.TxPvtReadWriteSet, error) {
	protoMsg := &rwset.TxPvtReadWriteSet{DataModel: rwset.TxReadWriteSet_KV}
	var nsProtoMsg *rwset.NsPvtReadWriteSet
	var err error
	for _, nsPvtRwSet := range txPvtRwSet.NsPvtRwSet {
		if nsProtoMsg, err = nsPvtRwSet.toProtoMsg(); err != nil {
			return nil, err
		}
		protoMsg.NsPvtRwset = append(protoMsg.NsPvtRwset, nsProtoMsg)
	}
	return protoMsg, nil
}

func txPvtRwSetFromProtoMsg(protoMsg *rwset.TxPvtReadWriteSet) (*TxPvtRwSet, error) {
	txPvtRwset := &TxPvtRwSet{}
	var nsPvtRwSet *NsPvtRwSet
	var err error
	for _, nsRwSetProtoMsg := range protoMsg.NsPvtRwset {
		if nsPvtRwSet, err = nsPvtRwSetFromProtoMsg(nsRwSetProtoMsg); err != nil {
			return nil, err
		}
		txPvtRwset.NsPvtRwSet = append(txPvtRwset.NsPvtRwSet, nsPvtRwSet)
	}
	return txPvtRwset, nil
}

func (nsPvtRwSet *NsPvtRwSet) toProtoMsg() (*rwset.NsPvtReadWriteSet, error) {
	protoMsg := &rwset.NsPvtReadWriteSet{Namespace: nsPvtRwSet.NameSpace}
	var err error
	var collPvtRwSetProtoMsg *rwset.CollectionPvtReadWriteSet
	for _, collPvtRwSet := range nsPvtRwSet.CollPvtRwSet {
		if collPvtRwSetProtoMsg, err = collPvtRwSet.toProtoMsg(); err != nil {
			return nil, err
		}
		protoMsg.CollectionPvtRwset = append(protoMsg.CollectionPvtRwset, collPvtRwSetProtoMsg)
	}
	return protoMsg, err
}

func nsPvtRwSetFromProtoMsg(protoMsg *rwset.NsPvtReadWriteSet) (*NsPvtRwSet, error) {
	nsPvtRwSet := &NsPvtRwSet{NameSpace: protoMsg.Namespace}
	for _, collPvtRwSetProtoMsg := range protoMsg.CollectionPvtRwset {
		var err error
		var collPvtRwSet *CollPvtRwSet
		if collPvtRwSet, err = collPvtRwSetFromProtoMsg(collPvtRwSetProtoMsg); err != nil {
			return nil, err
		}
		nsPvtRwSet.CollPvtRwSet = append(nsPvtRwSet.CollPvtRwSet, collPvtRwSet)
	}
	return nsPvtRwSet, nil
}

func (collPvtRwSet *CollPvtRwSet) toProtoMsg() (*rwset.CollectionPvtReadWriteSet, error) {
	var err error
	protoMsg := &rwset.CollectionPvtReadWriteSet{CollectionName: collPvtRwSet.CollectionName}
	if protoMsg.Rwset, err = proto.Marshal(collPvtRwSet.KvRwSet); err != nil {
		return nil, err
	}
	return protoMsg, nil
}

func collPvtRwSetFromProtoMsg(protoMsg *rwset.CollectionPvtReadWriteSet) (*CollPvtRwSet, error) {
	collPvtRwSet := &CollPvtRwSet{CollectionName: protoMsg.CollectionName, KvRwSet: &kvrwset.KVRWSet{}}
	if err := proto.Unmarshal(protoMsg.Rwset, collPvtRwSet.KvRwSet); err != nil {
		return nil, err
	}
	return collPvtRwSet, nil
}

// NewKVRead helps constructing proto message kvrwset.KVRead
func NewKVRead(key string, version *version.Height) *kvrwset.KVRead {
	return &kvrwset.KVRead{Key: key, Version: newProtoVersion(version)}
}

// NewVersion helps converting proto message kvrwset.Version to version.Height
func NewVersion(protoVersion *kvrwset.Version) *version.Height {
	if protoVersion == nil {
		return nil
	}
	return version.NewHeight(protoVersion.BlockNum, protoVersion.TxNum)
}

func newProtoVersion(height *version.Height) *kvrwset.Version {
	if height == nil {
		return nil
	}
	return &kvrwset.Version{BlockNum: height.BlockNum, TxNum: height.TxNum}
}

func newKVWrite(key string, value []byte) *kvrwset.KVWrite {
	return &kvrwset.KVWrite{Key: key, IsDelete: value == nil, Value: value}
}

func newPvtKVReadHash(key string, version *version.Height) (*kvrwset.KVReadHash, error) {
	keyHash, err := computeKeyHash(key)
	if err != nil {
		return nil, err
	}
	return &kvrwset.KVReadHash{KeyHash: keyHash, Version: newProtoVersion(version)}, nil
}

func newPvtKVWriteAndHash(key string, value []byte) (*kvrwset.KVWrite, *kvrwset.KVWriteHash, error) {
	kvWrite := newKVWrite(key, value)
	var keyHash, valueHash []byte
	var err error
	if keyHash, err = computeKeyHash(key); err != nil {
		return nil, nil, err
	}
	if !kvWrite.IsDelete {
		if valueHash, err = computeValueHash(value); err != nil {
			return nil, nil, err
		}
	}
	return kvWrite, &kvrwset.KVWriteHash{KeyHash: keyHash, IsDelete: kvWrite.IsDelete, ValueHash: valueHash}, nil
}

func computeKeyHash(input string) ([]byte, error) {
	return computeValueHash([]byte(input))
}

func computeValueHash(input []byte) ([]byte, error) {
	return bccspfactory.GetDefault().Hash([]byte(input), hashOpts)
}
