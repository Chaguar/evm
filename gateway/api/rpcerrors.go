/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package api

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hyperledger/fabric-x-evm/gateway/api/rpcerr"
	"github.com/hyperledger/fabric-x-evm/gateway/domain"
	"github.com/hyperledger/fabric-x-sdk/state"
)

// classifyValidationError maps a Backend.SendTransaction error to a typed
// JSON-RPC error. Internal lookup faults surface as -32603; everything else
// (geth txpool sentinels, domain.ErrUnprotectedTx) is a tx rejection (-32003).
func classifyValidationError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNonceLookup) {
		return rpcerr.Internal(err)
	}
	return rpcerr.TxRejected(err)
}

// classifyCallError maps a Backend.CallContract error to a typed JSON-RPC error:
// reverts (*domain.RevertError, with payload), non-revert execution failures
// (*domain.ExecutionError) and unavailable historical state surface as -32000;
// everything else is -32603.
func classifyCallError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, state.ErrHistoryUnavailable) {
		return rpcerr.HistoryUnavailable(historyDetail(err))
	}
	if revert, ok := errors.AsType[*domain.RevertError](err); ok {
		return rpcerr.ExecutionReverted(revert.Reason, hexutil.Encode(revert.Data))
	}
	if exec, ok := errors.AsType[*domain.ExecutionError](err); ok {
		return rpcerr.ExecutionError(exec.Message)
	}
	return rpcerr.Internal(err)
}

// classifyStateError maps a Backend state read error (eth_getBalance and
// friends): unavailable historical state is -32000 with the store's detail;
// anything else is returned unchanged.
func classifyStateError(err error) error {
	if errors.Is(err, state.ErrHistoryUnavailable) {
		return rpcerr.HistoryUnavailable(historyDetail(err))
	}
	return err
}

// historyDetail is what the store said beyond the sentinel's own text, e.g.
// "block 5 (retained history starts at 10, head is 20)". Empty when the store
// added nothing, as the query service does.
func historyDetail(err error) string {
	msg := err.Error()
	sentinel := state.ErrHistoryUnavailable.Error()
	if i := strings.Index(msg, sentinel); i >= 0 {
		msg = msg[i+len(sentinel):]
	}
	return strings.TrimLeft(msg, ": ")
}
