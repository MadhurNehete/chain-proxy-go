package process

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestMaterializeMRVDetailsFromLogs(t *testing.T) {
	t.Parallel()

	resetProxyMRVMetrics()
	result := materializeMRVDetails(map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "anchorReport",
			RcvAddr:  "erd1mrvregistry",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1mrvregistry",
						Identifier: "mrvReportAnchored",
						Topics: [][]byte{
							[]byte("report-001"),
							[]byte("tenant-public-001"),
							[]byte("farm-public-001"),
							[]byte("season-public-001"),
						},
						AdditionalData: [][]byte{
							[]byte("sha256:report-001"),
							[]byte("sha256"),
							[]byte("json-c14n-v1"),
							big.NewInt(2).Bytes(),
							big.NewInt(1710768000).Bytes(),
						},
					},
				},
			},
		},
	}, "")

	require.NotNil(t, result)
	require.True(t, result.IsMrv)
	require.True(t, result.HasAnchoredProof)
	require.Equal(t, "anchored", result.ProofStatus)
	require.Equal(t, "report-001", result.ReportID)
	require.Equal(t, "tenant-public-001", result.PublicTenantID)
	require.Equal(t, "farm-public-001", result.PublicFarmID)
	require.Equal(t, "season-public-001", result.PublicSeasonID)
	require.Equal(t, "sha256:report-001", result.ReportHash)
	require.Equal(t, "sha256", result.HashAlgo)
	require.Equal(t, "json-c14n-v1", result.Canonicalization)
	require.Equal(t, uint64(2), result.MethodologyVersion)
	require.Equal(t, uint64(1710768000), result.AnchoredAt)
	metrics := snapshotProxyMRVMetrics()
	require.Equal(t, uint64(1), metrics["mrv_signal_accepted"])
}

func TestMaterializeMRVDetailsFromV2Logs(t *testing.T) {
	t.Parallel()

	resetProxyMRVMetrics()
	result := materializeMRVDetails(map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "anchorReportV2",
			RcvAddr:  "erd1mrvregistry",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1mrvregistry",
						Identifier: "mrvReportAnchoredV2",
						Topics: [][]byte{
							[]byte("report-002"),
							[]byte("tenant-public-002"),
							[]byte("farm-public-002"),
							[]byte("season-public-002"),
						},
						AdditionalData: [][]byte{
							[]byte("sha3-256:report-002"),
							[]byte("sha3-256"),
							[]byte("rfc8785"),
							big.NewInt(2000).Bytes(),
							big.NewInt(1710769000).Bytes(),
							[]byte("KE-demo-project-002"),
							[]byte("sha3-256:evidence-manifest-002"),
						},
					},
				},
			},
		},
	}, "")

	require.NotNil(t, result)
	require.Equal(t, "report-002", result.ReportID)
	require.Equal(t, "KE-demo-project-002", result.PublicProjectID)
	require.Equal(t, "sha3-256:evidence-manifest-002", result.EvidenceManifestHash)
	require.Equal(t, uint64(2000), result.MethodologyVersion)
	require.Equal(t, uint64(1710769000), result.AnchoredAt)
}

func TestMaterializeMRVDetailsRejectsUnexpectedEmitter(t *testing.T) {
	t.Parallel()

	resetProxyMRVMetrics()
	result := materializeMRVDetails(map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "anchorReport",
			RcvAddr:  "erd1mrvregistry",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1spoofed",
						Identifier: "mrvReportAnchored",
					},
				},
			},
		},
	}, "")

	require.Nil(t, result)
	metrics := snapshotProxyMRVMetrics()
	require.Equal(t, uint64(1), metrics["mrv_signal_rejected"])
}

func TestMaterializeMRVDetailsIgnoresUnknownFunctions(t *testing.T) {
	t.Parallel()

	result := materializeMRVDetails(map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "customPing",
			RcvAddr:  "erd1mrvregistry",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1mrvregistry",
						Identifier: "mrvReportAnchored",
					},
				},
			},
		},
	}, "")

	require.Nil(t, result)
}
