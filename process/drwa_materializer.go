package process

import (
	"encoding/hex"
	"regexp"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-proxy-go/data"
)

// drwaCanonicalEvents is the exact set of event identifiers emitted by DRWA
// contracts and the protocol gate.  Prefix matching ("drwa*") is intentionally
// avoided: it would misclassify unrelated application events whose identifiers
// happen to start with "drwa", producing false-positive compliance signals.
var drwaCanonicalEvents = map[string]struct{}{
	"drwatokenpolicy":         {},
	"drwaassetregistered":     {},
	"drwaholdercompliance":    {},
	"drwatransferdenied":      {},
	"drwatransferallowed":     {},
	"drwaglobalpause":         {},
	"drwametadataprotection":  {},
	"drwaattestationrecorded": {},
}

func isDRWACanonicalEvent(identifier string) bool {
	if identifier == "" {
		return false
	}
	_, ok := drwaCanonicalEvents[strings.ToLower(identifier)]
	return ok
}

var drwaDenialPattern = regexp.MustCompile(`\bDRWA_[A-Z0-9_]+\b`)

const drwaDenialMetricLabelMaxLength = 64

func safeDRWADenialMetricLabel(identifier string) string {
	label := strings.ToLower(identifier)
	if len(label) > drwaDenialMetricLabelMaxLength {
		return label[:drwaDenialMetricLabelMaxLength]
	}

	return label
}

var drwaKnownFunctions = map[string]struct{}{
	"settokenpolicy":         {},
	"registerasset":          {},
	"syncholdercompliance":   {},
	"registeridentity":       {},
	"updatecompliancestatus": {},
	"recordattestation":      {},
	"manageddrwasyncmirror":  {},
	"drwa":                   {},
}

func isDRWAKnownFunction(function string) bool {
	if function == "" {
		return false
	}
	_, ok := drwaKnownFunctions[strings.ToLower(function)]
	return ok
}

func materializeDRWADetails(
	failReason string,
	scrs map[string]*transaction.ApiSmartContractResult,
	rootFunction string,
) *data.DrwaDetails {
	result := &data.DrwaDetails{}

	if denialIdentifier := extractDRWADenialCode(failReason); denialIdentifier != "" {
		recordProxyDRWAMetric("drwa_denial_detected")
		recordProxyDRWAMetric("drwa_denial_code_" + safeDRWADenialMetricLabel(denialIdentifier))
		result.IsDrwa = true
		result.DenialCode = denialIdentifier
		result.DenialMessage = failReason
	}

	for _, key := range sortedDRWASCRKeys(scrs) {
		scr := scrs[key]
		if scr == nil {
			continue
		}
		denialIdentifier := extractDRWADenialCode(scr.ReturnMessage)
		if denialIdentifier == "" {
			continue
		}

		recordProxyDRWAMetric("drwa_denial_detected")
		recordProxyDRWAMetric("drwa_denial_code_" + safeDRWADenialMetricLabel(denialIdentifier))
		result.IsDrwa = true
		if result.DenialCode == "" {
			result.DenialCode = denialIdentifier
			result.DenialMessage = scr.ReturnMessage
		}
	}

	for _, key := range sortedDRWASCRKeys(scrs) {
		scr := scrs[key]
		if scr == nil || scr.Logs == nil {
			continue
		}
		if !isDRWAKnownFunction(rootFunction) && !isDRWAKnownFunction(scr.Function) {
			continue
		}
		trustedEmitter := scr.RcvAddr
		if trustedEmitter == "" {
			trustedEmitter = scr.Logs.Address
		}

		for _, event := range scr.Logs.Events {
			if !isDRWACanonicalEvent(event.Identifier) {
				continue
			}
			if trustedEmitter == "" || event.Address != trustedEmitter {
				recordProxyDRWAMetric("drwa_signal_rejected")
				continue
			}

			recordProxyDRWAMetric("drwa_signal_accepted")
			result.IsDrwa = true
			result.HasComplianceSignal = true
			for _, topic := range event.Topics {
				if result.DenialCode == "" {
					if denialIdentifier := extractDRWADenialCode(string(topic)); denialIdentifier != "" {
						recordProxyDRWAMetric("drwa_denial_detected")
						recordProxyDRWAMetric("drwa_denial_code_" + safeDRWADenialMetricLabel(denialIdentifier))
						result.DenialCode = denialIdentifier
						result.DenialMessage = string(topic)
					}
				}
				result.DenialTopics = append(result.DenialTopics, hex.EncodeToString(topic))
			}
		}
	}

	if !result.IsDrwa && !result.HasComplianceSignal {
		return nil
	}

	return result
}

func sortedDRWASCRKeys(scrs map[string]*transaction.ApiSmartContractResult) []string {
	keys := make([]string, 0, len(scrs))
	for key := range scrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func extractDRWADenialCode(message string) string {
	match := drwaDenialPattern.FindString(message)
	if match == "" {
		return ""
	}

	return match
}
