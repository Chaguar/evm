/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package storage

import (
	"context"

	"github.com/hyperledger/fabric-x-evm/endorser/execution"
	"github.com/hyperledger/fabric-x-sdk/state"
	"github.com/hyperledger/fabric-x-sdk/state/queryservice"
)

// QueryServiceKVS reads committed state from the Fabric-X query service. It
// keeps no state of its own, so it only serves the latest state.
type QueryServiceKVS struct {
	*queryservice.Store
}

var _ KVS = QueryServiceKVS{}

// NewQueryServiceKVS returns a KVS reading through client.
func NewQueryServiceKVS(client queryservice.Client) QueryServiceKVS {
	return QueryServiceKVS{queryservice.NewStore(client)}
}

// NewSnapshot opens a view of the latest state. A non-nil blockNumber is
// passed to SnapshotAt, which returns state.ErrHistoryUnavailable: the query
// service has no history.
func (k QueryServiceKVS) NewSnapshot(blockNumber *uint64) (execution.ReadStore, error) {
	if blockNumber != nil {
		snap, err := k.SnapshotAt(context.Background(), *blockNumber)
		if err != nil {
			return nil, err
		}
		return queryServiceSnapshot{Snapshot: snap, height: *blockNumber}, nil
	}
	snap, err := k.Snapshot(context.Background())
	if err != nil {
		return nil, err
	}
	// Read the height after opening the view: the view is then no newer than it.
	height, err := k.BlockNumber(context.Background())
	if err != nil {
		snap.Close()
		return nil, err
	}
	return queryServiceSnapshot{Snapshot: snap, height: height}, nil
}

// queryServiceSnapshot is a query-service view with the height the store had
// seen when it was opened. The query service does not report a view's height,
// so this is the closest available.
type queryServiceSnapshot struct {
	state.Snapshot
	height uint64
}

// BlockNumber returns the height recorded when the view was opened.
func (s queryServiceSnapshot) BlockNumber() uint64 { return s.height }
