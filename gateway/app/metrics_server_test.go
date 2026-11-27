/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: LGPL-3.0-or-later
*/

package app

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-x-evm/gateway/metrics"
)

func TestMetricsHTTPServer_ServesPrometheusText(t *testing.T) {
	metrics.Default().TxReceived.Inc()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := newMetricsHTTPServer(ln.Addr().String())
	go func() {
		_ = srv.Serve(ln)
	}()
	t.Cleanup(func() { _ = srv.Close() })

	url := "http://" + ln.Addr().String() + "/metrics"
	var body string
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			b, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				t.Fatalf("read body: %v", readErr)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status=%d want 200", resp.StatusCode)
			}
			body = string(b)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s: %v", url, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(body, "evm_gateway_tx_received_total") {
		t.Fatalf("metrics body missing evm_gateway_tx_received_total:\n%s", body)
	}
}
