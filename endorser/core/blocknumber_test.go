/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package core

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric-x-evm/common"
	"github.com/hyperledger/fabric-x-evm/endorser/api"
	"github.com/hyperledger/fabric-x-sdk/endorsement"
)

// The gateway's number is used as-is, but it is required.
func TestExecute_BlockNumber(t *testing.T) {
	now := time.Now()
	tx := types.NewTx(&types.LegacyTx{Gas: 21000, GasPrice: big.NewInt(0)})
	newEndorser := func(eng *captureEngine) *Endorser {
		return &Endorser{
			Engine:    eng,
			builder:   &stubBuilder{resp: &peer.ProposalResponse{Response: &peer.Response{Status: common.StatusOK}}},
			maxFuture: time.Minute,
			maxPast:   time.Minute,
			now:       func() time.Time { return now },
		}
	}

	t.Run("passed to engine as-is", func(t *testing.T) {
		eng := &captureEngine{}
		resp, err := newEndorser(eng).Execute(context.Background(), endorsement.Invocation{}, tx, api.BlockEnv{Number: 1000, Time: now})
		if err != nil || resp.Response.Status != common.StatusOK {
			t.Fatalf("err=%v resp=%v", err, resp.GetResponse())
		}
		if eng.blockNumber != 1000 {
			t.Fatalf("blockNumber = %d, want 1000", eng.blockNumber)
		}
	})

	t.Run("missing", func(t *testing.T) {
		eng := &captureEngine{}
		resp, err := newEndorser(eng).Execute(context.Background(), endorsement.Invocation{}, tx, api.BlockEnv{Time: now})
		if err != nil {
			t.Fatal(err)
		}
		if resp.Response.Status != common.StatusTxRejected {
			t.Fatalf("status = %d, want StatusTxRejected", resp.Response.Status)
		}
		if eng.calls != 0 {
			t.Error("engine should not run without a block number")
		}
	})
}
