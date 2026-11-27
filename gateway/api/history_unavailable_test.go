/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hyperledger/fabric-x-evm/common"
	"github.com/hyperledger/fabric-x-evm/gateway/api/rpcerr"
	"github.com/hyperledger/fabric-x-sdk/state"
)

// Every method that reads state at a block reports a height the endorser
// doesn't hold as -32000, keeping the store's detail.
func TestHistoryUnavailable_PerMethod(t *testing.T) {
	ctx := context.Background()
	block := rpc.BlockNumberOrHashWithNumber(5)
	// As the gateway sees it: from a local store (wrapped), or rebuilt after a
	// hop (common.HistoryUnavailable).
	local := fmt.Errorf("%w: block 5 (retained history starts at 10, head is 20)", state.ErrHistoryUnavailable)
	remote := common.HistoryUnavailable(local.Error())
	const want = "historical state is not available: block 5 (retained history starts at 10, head is 20)"

	for _, src := range []struct {
		name string
		err  error
	}{{"local", local}, {"remote", remote}} {
		tests := []struct {
			name string
			call func(*EthAPI) error
			stub *stubBackend
		}{
			{"eth_getBalance", func(a *EthAPI) error { _, err := a.GetBalance(ctx, testAddr, block); return err }, &stubBackend{balanceErr: src.err}},
			{"eth_getCode", func(a *EthAPI) error { _, err := a.GetCode(ctx, testAddr, block); return err }, &stubBackend{codeErr: src.err}},
			{"eth_getStorageAt", func(a *EthAPI) error { _, err := a.GetStorageAt(ctx, testAddr, "0x0", block); return err }, &stubBackend{storageErr: src.err}},
			{"eth_getTransactionCount", func(a *EthAPI) error { _, err := a.GetTransactionCount(ctx, testAddr, block); return err }, &stubBackend{nonceErr: src.err}},
			{"eth_call", func(a *EthAPI) error { _, err := a.Call(ctx, map[string]any{}, block); return err }, &stubBackend{callErr: src.err}},
			{"eth_estimateGas", func(a *EthAPI) error { _, err := a.EstimateGas(ctx, map[string]any{}, &block); return err }, &stubBackend{estErr: src.err}},
		}
		for _, tt := range tests {
			t.Run(src.name+"/"+tt.name, func(t *testing.T) {
				err := tt.call(NewEthAPI(tt.stub))
				rpcErr, ok := errors.AsType[rpc.Error](err)
				if !ok {
					t.Fatalf("err = %T (%v), want an rpc.Error", err, err)
				}
				if rpcErr.ErrorCode() != rpcerr.CodeServerError {
					t.Errorf("code = %d, want %d", rpcErr.ErrorCode(), rpcerr.CodeServerError)
				}
				if err.Error() != want {
					t.Errorf("message = %q, want %q", err.Error(), want)
				}
			})
		}
	}
}

// The query service returns the bare sentinel: the message is the plain one.
func TestHistoryUnavailable_NoDetail(t *testing.T) {
	err := classifyCallError(state.ErrHistoryUnavailable)
	if err.Error() != "historical state is not available" {
		t.Errorf("message = %q", err.Error())
	}
}

// Other state read errors are left as they were.
func TestClassifyStateError_PassesOtherErrorsThrough(t *testing.T) {
	if got := classifyStateError(errBoom); got != errBoom {
		t.Errorf("got %v, want errBoom unchanged", got)
	}
}
