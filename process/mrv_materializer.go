package process

import (
	"math/big"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-proxy-go/data"
)

var mrvCanonicalEvents = map[string]struct{}{
	"mrvreportanchored":   {},
	"mrvreportanchoredv2": {},
	"mrvproofmaterialized": {},
}

func isMRVCanonicalEvent(identifier string) bool {
	if identifier == "" {
		return false
	}

	_, ok := mrvCanonicalEvents[strings.ToLower(identifier)]
	return ok
}

var mrvKnownFunctions = map[string]struct{}{
	"anchorreport":          {},
	"anchorreportv2":        {},
	"getreportproof":        {},
	"getreportproofbyseason": {},
}

func isMRVKnownFunction(function string) bool {
	if function == "" {
		return false
	}

	_, ok := mrvKnownFunctions[strings.ToLower(function)]
	return ok
}

func materializeMRVDetails(
	scrs map[string]*transaction.ApiSmartContractResult,
	rootFunction string,
) *data.MrvDetails {
	result := &data.MrvDetails{}

	for _, key := range sortedMRVSCRKeys(scrs) {
		scr := scrs[key]
		if scr == nil || scr.Logs == nil {
			continue
		}
		if !isMRVKnownFunction(rootFunction) && !isMRVKnownFunction(scr.Function) {
			continue
		}
		processMRVSCR(scr, result)
	}

	if !result.IsMrv {
		return nil
	}

	return result
}

func processMRVSCR(scr *transaction.ApiSmartContractResult, result *data.MrvDetails) {
	trustedEmitter := scr.RcvAddr
	if trustedEmitter == "" {
		trustedEmitter = scr.Logs.Address
	}

	for _, event := range scr.Logs.Events {
		if !isMRVCanonicalEvent(event.Identifier) {
			continue
		}
		if trustedEmitter == "" || event.Address != trustedEmitter {
			recordProxyMRVMetric("mrv_signal_rejected")
			continue
		}
		recordProxyMRVMetric("mrv_signal_accepted")
		applyMRVEventToResult(event, result)
	}
}

func applyMRVEventToResult(event *transaction.Events, result *data.MrvDetails) {
	result.IsMrv = true
	result.SourceEvent = event.Identifier
	result.HasAnchoredProof = true
	result.ProofStatus = "anchored"

	applyMRVTopics(event.Topics, result)
	applyMRVAdditionalData(event.AdditionalData, result)
}

func applyMRVTopics(topics [][]byte, result *data.MrvDetails) {
	if len(topics) > 0 {
		result.ReportID = string(topics[0])
	}
	if len(topics) > 1 {
		result.PublicTenantID = string(topics[1])
	}
	if len(topics) > 2 {
		result.PublicFarmID = string(topics[2])
	}
	if len(topics) > 3 {
		result.PublicSeasonID = string(topics[3])
	}
}

func applyMRVAdditionalData(data [][]byte, result *data.MrvDetails) {
	if len(data) > 0 {
		result.ReportHash = string(data[0])
	}
	if len(data) > 1 {
		result.HashAlgo = string(data[1])
	}
	if len(data) > 2 {
		result.Canonicalization = string(data[2])
	}
	if len(data) > 3 {
		result.MethodologyVersion = big.NewInt(0).SetBytes(data[3]).Uint64()
	}
	if len(data) > 4 {
		result.AnchoredAt = big.NewInt(0).SetBytes(data[4]).Uint64()
	}
	if len(data) > 5 {
		result.PublicProjectID = string(data[5])
	}
	if len(data) > 6 {
		result.EvidenceManifestHash = string(data[6])
	}
}

func sortedMRVSCRKeys(scrs map[string]*transaction.ApiSmartContractResult) []string {
	keys := make([]string, 0, len(scrs))
	for key := range scrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
