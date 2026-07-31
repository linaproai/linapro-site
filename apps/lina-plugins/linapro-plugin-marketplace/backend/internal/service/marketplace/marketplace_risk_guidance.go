// This file implements marketplace risk finding disposition, blocking policy,
// bounded evidence packaging, and presentation ordering for risk list APIs.

package marketplace

import (
	"sort"
	"strings"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
)

// Maximum evidence list items persisted on one risk payload.
const marketplaceRiskEvidenceLimit = 20

// PackageDiagnosticEvidence carries bounded scanner evidence for one finding.
type PackageDiagnosticEvidence struct {
	Files         []string
	Services      []DiagnosticServiceEvidence
	Routes        []DiagnosticRouteEvidence
	ExpectedPath  string
	ExpectedField string
	Example       string
	TotalCount    int
	Truncated     bool
}

// DiagnosticServiceEvidence summarizes one host service request.
type DiagnosticServiceEvidence struct {
	Service string   `json:"service"`
	Methods []string `json:"methods,omitempty"`
	Tables  []string `json:"tables,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	Keys    []string `json:"keys,omitempty"`
}

// DiagnosticRouteEvidence summarizes one dynamic route declaration.
type DiagnosticRouteEvidence struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Permission string `json:"permission,omitempty"`
	Access     string `json:"access,omitempty"`
}

// packageDiagnosticRiskPayload is the JSON shape stored on risk rows and returned
// to clients (plus top-level disposition/blocking on MarketplaceRiskItem).
type packageDiagnosticRiskPayload struct {
	Code          string                      `json:"code"`
	Disposition   string                      `json:"disposition,omitempty"`
	Blocking      bool                        `json:"blocking,omitempty"`
	Files         []string                    `json:"files,omitempty"`
	Services      []DiagnosticServiceEvidence `json:"services,omitempty"`
	Routes        []DiagnosticRouteEvidence   `json:"routes,omitempty"`
	ExpectedPath  string                      `json:"expectedPath,omitempty"`
	ExpectedField string                      `json:"expectedField,omitempty"`
	Example       string                      `json:"example,omitempty"`
	TotalCount    int                         `json:"totalCount,omitempty"`
	Truncated     bool                        `json:"truncated,omitempty"`
}

// riskDispositionPolicy describes how a stable scanner code should be treated.
type riskDispositionPolicy struct {
	Disposition marketv1.MarketplaceRiskDisposition
	Blocking    bool
}

// riskDispositionPolicyByCode maps known scanner codes to disposition/blocking.
// Unknown codes default to need_attention / non-blocking.
// Framework compatibility declaration is disclosure-only (need_attention) and
// never blocks review submission.
var riskDispositionPolicyByCode = map[string]riskDispositionPolicy{
	"framework_dependency_missing": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"i18n_files_missing": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedFix,
		Blocking:    true,
	},
	"dynamic_manifest_resources_missing": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedFix,
		Blocking:    true,
	},
	"source_sql_present": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"dynamic_sql_present": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"dynamic_mock_sql_present": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"dynamic_host_services_present": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"dynamic_routes_present": {
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	},
	"source_docs_indexed": {
		Disposition: marketv1.MarketplaceRiskDispositionInfoOnly,
		Blocking:    false,
	},
	"dynamic_runtime_detected": {
		Disposition: marketv1.MarketplaceRiskDispositionInfoOnly,
		Blocking:    false,
	},
}

// resolveRiskDispositionPolicy returns disposition and blocking for one code.
func resolveRiskDispositionPolicy(code string) riskDispositionPolicy {
	normalized := strings.TrimSpace(code)
	if policy, ok := riskDispositionPolicyByCode[normalized]; ok {
		return policy
	}
	return riskDispositionPolicy{
		Disposition: marketv1.MarketplaceRiskDispositionNeedAttention,
		Blocking:    false,
	}
}

// buildPackageDiagnosticRiskPayload serializes one diagnostic for risk storage.
func buildPackageDiagnosticRiskPayload(diagnostic *PackageDiagnostic) packageDiagnosticRiskPayload {
	if diagnostic == nil {
		return packageDiagnosticRiskPayload{}
	}
	code := strings.TrimSpace(diagnostic.Code)
	policy := resolveRiskDispositionPolicy(code)
	payload := packageDiagnosticRiskPayload{
		Code:        code,
		Disposition: policy.Disposition.String(),
		Blocking:    policy.Blocking,
	}
	if diagnostic.Evidence == nil {
		return payload
	}
	ev := diagnostic.Evidence
	payload.Files = cloneStringSlice(ev.Files)
	payload.Services = cloneServiceEvidence(ev.Services)
	payload.Routes = cloneRouteEvidence(ev.Routes)
	payload.ExpectedPath = strings.TrimSpace(ev.ExpectedPath)
	payload.ExpectedField = strings.TrimSpace(ev.ExpectedField)
	payload.Example = strings.TrimSpace(ev.Example)
	payload.TotalCount = ev.TotalCount
	payload.Truncated = ev.Truncated
	return payload
}

// boundStringEvidence truncates a path/name list to the evidence limit.
func boundStringEvidence(items []string) (result []string, total int, truncated bool) {
	total = len(items)
	if total == 0 {
		return nil, 0, false
	}
	if total <= marketplaceRiskEvidenceLimit {
		return cloneStringSlice(items), total, false
	}
	return cloneStringSlice(items[:marketplaceRiskEvidenceLimit]), total, true
}

// boundServiceEvidence truncates host service evidence.
func boundServiceEvidence(items []DiagnosticServiceEvidence) (result []DiagnosticServiceEvidence, total int, truncated bool) {
	total = len(items)
	if total == 0 {
		return nil, 0, false
	}
	if total <= marketplaceRiskEvidenceLimit {
		return cloneServiceEvidence(items), total, false
	}
	return cloneServiceEvidence(items[:marketplaceRiskEvidenceLimit]), total, true
}

// boundRouteEvidence truncates dynamic route evidence.
func boundRouteEvidence(items []DiagnosticRouteEvidence) (result []DiagnosticRouteEvidence, total int, truncated bool) {
	total = len(items)
	if total == 0 {
		return nil, 0, false
	}
	if total <= marketplaceRiskEvidenceLimit {
		return cloneRouteEvidence(items), total, false
	}
	return cloneRouteEvidence(items[:marketplaceRiskEvidenceLimit]), total, true
}

func cloneStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}

func cloneServiceEvidence(items []DiagnosticServiceEvidence) []DiagnosticServiceEvidence {
	if len(items) == 0 {
		return nil
	}
	out := make([]DiagnosticServiceEvidence, len(items))
	for i, item := range items {
		out[i] = DiagnosticServiceEvidence{
			Service: item.Service,
			Methods: cloneStringSlice(item.Methods),
			Tables:  cloneStringSlice(item.Tables),
			Paths:   cloneStringSlice(item.Paths),
			Keys:    cloneStringSlice(item.Keys),
		}
	}
	return out
}

func cloneRouteEvidence(items []DiagnosticRouteEvidence) []DiagnosticRouteEvidence {
	if len(items) == 0 {
		return nil
	}
	out := make([]DiagnosticRouteEvidence, len(items))
	copy(out, items)
	return out
}

// applyRiskGuidanceToItem fills disposition/blocking on a projected risk item.
// The code policy table is authoritative so strategy changes apply to already
// stored risk rows (including legacy payloads that still embed old blocking flags).
func applyRiskGuidanceToItem(item *marketv1.MarketplaceRiskItem) {
	if item == nil {
		return
	}
	code := riskCodeFromPayload(item.Payload)
	policy := resolveRiskDispositionPolicy(code)
	item.Disposition = policy.Disposition
	item.Blocking = policy.Blocking
}

func riskCodeFromPayload(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload["code"]
	if !ok {
		return ""
	}
	text, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

// riskDispositionRank lower values sort first: need_fix → need_attention → info_only.
func riskDispositionRank(disposition marketv1.MarketplaceRiskDisposition) int {
	switch marketv1.MarketplaceRiskDisposition(strings.ToLower(strings.TrimSpace(disposition.String()))) {
	case marketv1.MarketplaceRiskDispositionNeedFix:
		return 0
	case marketv1.MarketplaceRiskDispositionNeedAttention:
		return 1
	case marketv1.MarketplaceRiskDispositionInfoOnly:
		return 2
	default:
		return 3
	}
}

// sortMarketplaceRiskItems orders findings for publisher/reviewer workbenches:
// blocking first, then disposition, then severity (high → warning → info).
func sortMarketplaceRiskItems(items []*marketv1.MarketplaceRiskItem) {
	if len(items) < 2 {
		return
	}
	for _, item := range items {
		applyRiskGuidanceToItem(item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left == nil && right == nil {
			return false
		}
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		if left.Blocking != right.Blocking {
			return left.Blocking
		}
		if leftRank, rightRank := riskDispositionRank(left.Disposition), riskDispositionRank(right.Disposition); leftRank != rightRank {
			return leftRank < rightRank
		}
		return riskSeverityRank(left.Severity) < riskSeverityRank(right.Severity)
	})
}

// sortMarketplaceRiskItemsBySeverity keeps the historical name as an alias that
// now applies the full guidance order (blocking → disposition → severity).
func sortMarketplaceRiskItemsBySeverity(items []*marketv1.MarketplaceRiskItem) {
	sortMarketplaceRiskItems(items)
}
