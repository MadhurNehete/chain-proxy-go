package process

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestMaterializeDRWADetailsFromFailReason(t *testing.T) {

	resetProxyDRWAMetrics()
	result := materializeDRWADetails("execution failed: DRWA_KYC_REQUIRED receiver not eligible", nil, "")
	require.NotNil(t, result)
	require.True(t, result.IsDrwa)
	require.Equal(t, "DRWA_KYC_REQUIRED", result.DenialCode)
	metrics := snapshotProxyDRWAMetrics()
	require.Equal(t, uint64(1), metrics["drwa_denial_detected"])
	require.Equal(t, uint64(1), metrics["drwa_denial_code_drwa_kyc_required"])
}

func TestMaterializeDRWADetailsFromSCRReturnMessage(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {ReturnMessage: "DRWA_TOKEN_PAUSED token is paused"},
	}, "")
	require.NotNil(t, result)
	require.Equal(t, "DRWA_TOKEN_PAUSED", result.DenialCode)
}

func TestMaterializeDRWADetailsFromLogs(t *testing.T) {

	resetProxyDRWAMetrics()
	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "setTokenPolicy",
			RcvAddr:  "erd1policy",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1policy",
						Identifier: "drwaTransferDenied",
						Topics:     [][]byte{[]byte("DRWA_JURISDICTION_BLOCKED")},
					},
				},
			},
		},
	}, "")
	require.NotNil(t, result)
	require.True(t, result.HasComplianceSignal)
	require.Len(t, result.DenialTopics, 1)
	metrics := snapshotProxyDRWAMetrics()
	require.Equal(t, uint64(1), metrics["drwa_signal_accepted"])
}

func TestMaterializeDRWADetailsIgnoresTransportFailures(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("gateway timeout while fetching SCR", map[string]*transaction.ApiSmartContractResult{
		"scr": {ReturnMessage: "upstream unavailable"},
	}, "")
	require.Nil(t, result)
}

func TestMaterializeDRWADetailsPrefersFailReasonOverSCRDenial(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("execution failed: DRWA_KYC_REQUIRED source holder missing approval", map[string]*transaction.ApiSmartContractResult{
		"scr": {ReturnMessage: "DRWA_TOKEN_PAUSED token is paused"},
	}, "")

	require.NotNil(t, result)
	require.Equal(t, "DRWA_KYC_REQUIRED", result.DenialCode)
	require.Contains(t, result.DenialMessage, "DRWA_KYC_REQUIRED")
}

func TestMaterializeDRWADetailsUsesDeterministicSCROrder(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"z-scr": {ReturnMessage: "DRWA_TOKEN_PAUSED token is paused"},
		"a-scr": {ReturnMessage: "DRWA_KYC_REQUIRED holder is not eligible"},
	}, "")

	require.NotNil(t, result)
	require.Equal(t, "DRWA_KYC_REQUIRED", result.DenialCode)
	require.Contains(t, result.DenialMessage, "DRWA_KYC_REQUIRED")
}

func TestMaterializeDRWADetailsIgnoresSubstringFalsePositive(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("execution failed: XDRWA_KYC_REQUIREDY", nil, "")
	require.Nil(t, result)
}

func TestMaterializeDRWADetailsIgnoresSubstringFalsePositiveInTopic(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "recordAttestation",
			RcvAddr:  "erd1attestation",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1attestation",
						Identifier: "drwaAttestationRecorded",
						Topics:     [][]byte{[]byte("XDRWA_GLOBAL_PAUSEY")},
					},
				},
			},
		},
	}, "")

	require.NotNil(t, result)
	require.True(t, result.IsDrwa)
	require.True(t, result.HasComplianceSignal)
	require.Equal(t, "", result.DenialCode)
}

func TestMaterializeDRWADetailsExtractsDenialFromTopic(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "setTokenPolicy",
			RcvAddr:  "erd1policy",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1policy",
						Identifier: "drwaTransferDenied",
						Topics:     [][]byte{[]byte("DRWA_GLOBAL_PAUSE holder erd1xyz missing attestation")},
					},
				},
			},
		},
	}, "")

	require.NotNil(t, result)
	require.Equal(t, "DRWA_GLOBAL_PAUSE", result.DenialCode)
	require.Equal(t, "DRWA_GLOBAL_PAUSE holder erd1xyz missing attestation", result.DenialMessage)
}

func TestMaterializeDRWADetailsRecognizesAttestationSignalWithoutDenial(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "recordAttestation",
			RcvAddr:  "erd1attestation",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1attestation",
						Identifier: "drwaAttestationRecorded",
						Topics:     [][]byte{[]byte("HOTEL-1234"), []byte("erd1subject")},
					},
				},
			},
		},
	}, "")

	require.NotNil(t, result)
	require.True(t, result.HasComplianceSignal)
	require.Equal(t, "", result.DenialCode)
}

func TestMaterializeDRWADetailsIgnoresCanonicalEventsWithoutTrustedFunctionContext(t *testing.T) {
	t.Parallel()

	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "customPing",
			RcvAddr:  "erd1spoof",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1spoof",
						Identifier: "drwaTransferAllowed",
						Topics:     [][]byte{[]byte("DRWA_TRANSFER_LOCKED")},
					},
				},
			},
		},
	}, "")

	require.Nil(t, result)
}

func TestMaterializeDRWADetailsIgnoresCanonicalEventsFromUnexpectedEmitter(t *testing.T) {

	resetProxyDRWAMetrics()
	result := materializeDRWADetails("", map[string]*transaction.ApiSmartContractResult{
		"scr": {
			Function: "setTokenPolicy",
			RcvAddr:  "erd1policy",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Address:    "erd1spoof",
						Identifier: "drwaTransferAllowed",
						Topics:     [][]byte{[]byte("DRWA_TRANSFER_LOCKED")},
					},
				},
			},
		},
	}, "")

	require.Nil(t, result)
	metrics := snapshotProxyDRWAMetrics()
	require.Equal(t, uint64(1), metrics["drwa_signal_rejected"])
}

func TestMaterializeDRWADetailsTruncatesLongDenialMetricKeys(t *testing.T) {

	resetProxyDRWAMetrics()
	longIdentifier := "DRWA_" + strings.Repeat("A", 200)
	result := materializeDRWADetails("execution failed: "+longIdentifier+" reason", nil, "")

	require.NotNil(t, result)
	require.Equal(t, longIdentifier, result.DenialCode)
	metrics := snapshotProxyDRWAMetrics()
	require.Equal(t, uint64(1), metrics["drwa_denial_detected"])
	require.Equal(t, uint64(1), metrics["drwa_denial_code_"+strings.ToLower(longIdentifier[:64])])
}
