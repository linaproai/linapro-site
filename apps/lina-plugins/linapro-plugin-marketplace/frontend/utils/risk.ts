/**
 * Marketplace risk finding display helpers.
 *
 * Scanner rows persist English source text in `summary` and a stable diagnostic
 * code in `payload.code`. UI locales render the code through plugin i18n keys
 * (title, reason, remediation, acceptance) and sort by blocking → disposition
 * → severity so publishers see fixable items first.
 */

export type MarketplaceRiskSeverityRank = "high" | "info" | "warning" | string;

export type MarketplaceRiskDisposition =
  | "info_only"
  | "need_attention"
  | "need_fix"
  | string;

export type MarketplaceRiskFindingLike = {
  blocking?: boolean | null;
  disposition?: MarketplaceRiskDisposition | null;
  payload?: null | Record<string, unknown>;
  severity?: MarketplaceRiskSeverityRank | null;
  summary?: null | string;
};

/** Runtime i18n translate function used by Vue pages (`t` from useI18n). */
export type MarketplaceRiskTranslate = (
  key: string,
  params?: Record<string, unknown>,
) => string;

/** Optional `te` existence check from vue-i18n; when omitted, key equality is used. */
export type MarketplaceRiskTranslateExists = (key: string) => boolean;

const riskFindingKeyPrefix =
  "plugin.linapro-plugin-marketplace.detail.riskFinding.";

const riskDispositionKeyPrefix =
  "plugin.linapro-plugin-marketplace.detail.riskDisposition.";

/**
 * Scanner diagnostic codes that have dedicated runtime i18n entries under
 * `detail.riskFinding.*`. Keep this list in sync with backend package scanners.
 */
export const MARKETPLACE_RISK_FINDING_CODES = [
  "source_sql_present",
  "source_docs_indexed",
  "framework_dependency_missing",
  "i18n_files_missing",
  "dynamic_runtime_detected",
  "dynamic_host_services_present",
  "dynamic_routes_present",
  "dynamic_sql_present",
  "dynamic_mock_sql_present",
  "dynamic_manifest_resources_missing",
] as const;

export type MarketplaceRiskFindingCode =
  (typeof MARKETPLACE_RISK_FINDING_CODES)[number];

/** Structured guidance suffixes for each finding code. */
export const MARKETPLACE_RISK_GUIDANCE_SUFFIXES = [
  "reason",
  "remediation",
  "acceptance",
] as const;

/**
 * Full runtime i18n keys for every known scanner finding code + guidance fields.
 * Title uses `.title` so nested runtime message trees do not overwrite the
 * leaf title when sibling keys like `.reason` exist under the same code path.
 */
export const MARKETPLACE_RISK_FINDING_I18N_KEYS = [
  ...MARKETPLACE_RISK_FINDING_CODES.map(
    (code) =>
      `plugin.linapro-plugin-marketplace.detail.riskFinding.${code}.title` as const,
  ),
  ...MARKETPLACE_RISK_FINDING_CODES.flatMap((code) =>
    MARKETPLACE_RISK_GUIDANCE_SUFFIXES.map(
      (suffix) =>
        `plugin.linapro-plugin-marketplace.detail.riskFinding.${code}.${suffix}` as const,
    ),
  ),
];

/** Frontend fallback policy when API omits disposition/blocking (legacy rows). */
const RISK_DISPOSITION_POLICY: Record<
  string,
  { blocking: boolean; disposition: MarketplaceRiskDisposition }
> = {
  // Framework compatibility is disclosure-only; never blocks review submit.
  framework_dependency_missing: {
    blocking: false,
    disposition: "need_attention",
  },
  i18n_files_missing: { blocking: true, disposition: "need_fix" },
  dynamic_manifest_resources_missing: {
    blocking: true,
    disposition: "need_fix",
  },
  source_sql_present: { blocking: false, disposition: "need_attention" },
  dynamic_sql_present: { blocking: false, disposition: "need_attention" },
  dynamic_mock_sql_present: { blocking: false, disposition: "need_attention" },
  dynamic_host_services_present: {
    blocking: false,
    disposition: "need_attention",
  },
  dynamic_routes_present: { blocking: false, disposition: "need_attention" },
  source_docs_indexed: { blocking: false, disposition: "info_only" },
  dynamic_runtime_detected: { blocking: false, disposition: "info_only" },
};

/**
 * Severity sort rank: lower value means higher priority in the risk list.
 * high (0) → warning (1) → info (2) → unknown (3).
 */
export function marketplaceRiskSeverityRank(
  severity: MarketplaceRiskSeverityRank | null | undefined,
): number {
  switch ((severity || "").trim().toLowerCase()) {
    case "high": {
      return 0;
    }
    case "warning": {
      return 1;
    }
    case "info": {
      return 2;
    }
    default: {
      return 3;
    }
  }
}

/**
 * Disposition sort rank: need_fix first, then need_attention, then info_only.
 */
export function marketplaceRiskDispositionRank(
  disposition: MarketplaceRiskDisposition | null | undefined,
): number {
  switch ((disposition || "").trim().toLowerCase()) {
    case "need_fix": {
      return 0;
    }
    case "need_attention": {
      return 1;
    }
    case "info_only": {
      return 2;
    }
    default: {
      return 3;
    }
  }
}

/**
 * Resolve the stable scanner diagnostic code stored on a risk finding payload.
 */
export function marketplaceRiskFindingCode(
  risk: MarketplaceRiskFindingLike | null | undefined,
): string {
  const code = risk?.payload?.code;
  if (typeof code !== "string") {
    return "";
  }
  return code.trim();
}

/**
 * Resolve disposition from the stable scanner code policy table.
 * Policy is authoritative so strategy changes apply even when API/payload still
 * embed older disposition values from a previous scan.
 */
export function marketplaceRiskDisposition(
  risk: MarketplaceRiskFindingLike | null | undefined,
): MarketplaceRiskDisposition {
  const code = marketplaceRiskFindingCode(risk);
  if (code && RISK_DISPOSITION_POLICY[code]) {
    return RISK_DISPOSITION_POLICY[code].disposition;
  }
  const direct = (risk?.disposition || "").trim();
  if (
    direct === "need_fix" ||
    direct === "need_attention" ||
    direct === "info_only"
  ) {
    return direct;
  }
  const fromPayload = risk?.payload?.disposition;
  if (
    typeof fromPayload === "string" &&
    (fromPayload === "need_fix" ||
      fromPayload === "need_attention" ||
      fromPayload === "info_only")
  ) {
    return fromPayload;
  }
  return "need_attention";
}

/**
 * Resolve blocking flag from the stable scanner code policy table.
 * Known codes always use policy (ignores stale payload.blocking=true).
 */
export function marketplaceRiskBlocking(
  risk: MarketplaceRiskFindingLike | null | undefined,
): boolean {
  const code = marketplaceRiskFindingCode(risk);
  if (code && RISK_DISPOSITION_POLICY[code]) {
    return RISK_DISPOSITION_POLICY[code].blocking === true;
  }
  if (risk?.blocking === true) {
    return true;
  }
  if (risk?.blocking === false) {
    return false;
  }
  if (risk?.payload?.blocking === true) {
    return true;
  }
  if (risk?.payload?.blocking === false) {
    return false;
  }
  return false;
}

/**
 * Stable sort: blocking first → disposition → severity. Same-rank items keep order.
 */
export function sortMarketplaceRiskFindingsBySeverity<
  T extends MarketplaceRiskFindingLike,
>(items: readonly T[] | null | undefined): T[] {
  if (!items || items.length === 0) {
    return [];
  }
  return items
    .map((item, index) => ({ index, item }))
    .sort((left, right) => {
      const leftBlocking = marketplaceRiskBlocking(left.item) ? 0 : 1;
      const rightBlocking = marketplaceRiskBlocking(right.item) ? 0 : 1;
      if (leftBlocking !== rightBlocking) {
        return leftBlocking - rightBlocking;
      }
      const dispositionDiff =
        marketplaceRiskDispositionRank(
          marketplaceRiskDisposition(left.item),
        ) -
        marketplaceRiskDispositionRank(
          marketplaceRiskDisposition(right.item),
        );
      if (dispositionDiff !== 0) {
        return dispositionDiff;
      }
      const severityDiff =
        marketplaceRiskSeverityRank(left.item.severity) -
        marketplaceRiskSeverityRank(right.item.severity);
      if (severityDiff !== 0) {
        return severityDiff;
      }
      return left.index - right.index;
    })
    .map((entry) => entry.item);
}

/**
 * Build the runtime i18n key base for one scanner diagnostic code
 * (`detail.riskFinding.<code>`). Guidance fields append `.reason` etc.; the
 * title key appends `.title`.
 */
export function marketplaceRiskFindingBaseKey(code: string): string {
  const normalized = code.trim();
  if (!normalized) {
    return "";
  }
  return `${riskFindingKeyPrefix}${normalized}`;
}

/**
 * Build the runtime i18n key for one scanner diagnostic code title.
 * Uses `.title` suffix so nestFlatMessageMap does not replace the title leaf
 * with a map when reason/remediation/acceptance siblings are present.
 */
export function marketplaceRiskFindingMessageKey(code: string): string {
  const base = marketplaceRiskFindingBaseKey(code);
  if (!base) {
    return "";
  }
  return `${base}.title`;
}

function translateKey(
  t: MarketplaceRiskTranslate,
  key: string,
  te?: MarketplaceRiskTranslateExists,
): string {
  if (!key) {
    return "";
  }
  if (typeof te === "function") {
    if (!te(key)) {
      return "";
    }
    return t(key);
  }
  const translated = t(key);
  if (!translated || translated === key) {
    return "";
  }
  return translated;
}

/**
 * Localize one risk finding summary for the current UI locale.
 */
export function formatMarketplaceRiskFindingSummary(
  t: MarketplaceRiskTranslate,
  risk: MarketplaceRiskFindingLike | null | undefined,
  te?: MarketplaceRiskTranslateExists,
): string {
  const fallback = (risk?.summary || "").trim();
  const code = marketplaceRiskFindingCode(risk);
  if (!code) {
    return fallback;
  }
  const key = marketplaceRiskFindingMessageKey(code);
  const translated = translateKey(t, key, te);
  return translated || fallback;
}

/**
 * Localize structured guidance field (reason / remediation / acceptance).
 */
export function formatMarketplaceRiskFindingGuidance(
  t: MarketplaceRiskTranslate,
  risk: MarketplaceRiskFindingLike | null | undefined,
  field: "acceptance" | "reason" | "remediation",
  te?: MarketplaceRiskTranslateExists,
): string {
  const code = marketplaceRiskFindingCode(risk);
  if (!code) {
    return "";
  }
  const base = marketplaceRiskFindingBaseKey(code);
  if (!base) {
    return "";
  }
  return translateKey(t, `${base}.${field}`, te);
}

/**
 * Localize disposition label.
 */
export function formatMarketplaceRiskDisposition(
  t: MarketplaceRiskTranslate,
  disposition: MarketplaceRiskDisposition | null | undefined,
): string {
  const normalized = (disposition || "").trim().toLowerCase();
  if (!normalized) {
    return "";
  }
  const key = `${riskDispositionKeyPrefix}${normalized}`;
  const translated = t(key);
  if (!translated || translated === key) {
    return normalized;
  }
  return translated;
}

export type MarketplaceRiskServiceEvidence = {
  keys?: string[];
  methods?: string[];
  paths?: string[];
  service?: string;
  tables?: string[];
};

export type MarketplaceRiskRouteEvidence = {
  access?: string;
  method?: string;
  path?: string;
  permission?: string;
};

export type MarketplaceRiskEvidence = {
  example: string;
  expectedField: string;
  expectedPath: string;
  files: string[];
  routes: MarketplaceRiskRouteEvidence[];
  services: MarketplaceRiskServiceEvidence[];
  totalCount: number;
  truncated: boolean;
};

function asStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item): item is string => typeof item === "string")
    .map((item) => item.trim())
    .filter(Boolean);
}

function asServiceEvidence(value: unknown): MarketplaceRiskServiceEvidence[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item): item is Record<string, unknown> => !!item && typeof item === "object")
    .map((item) => ({
      keys: asStringArray(item.keys),
      methods: asStringArray(item.methods),
      paths: asStringArray(item.paths),
      service: typeof item.service === "string" ? item.service : "",
      tables: asStringArray(item.tables),
    }))
    .filter((item) => item.service);
}

function asRouteEvidence(value: unknown): MarketplaceRiskRouteEvidence[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((item): item is Record<string, unknown> => !!item && typeof item === "object")
    .map((item) => ({
      access: typeof item.access === "string" ? item.access : "",
      method: typeof item.method === "string" ? item.method : "",
      path: typeof item.path === "string" ? item.path : "",
      permission: typeof item.permission === "string" ? item.permission : "",
    }))
    .filter((item) => item.method || item.path);
}

/**
 * Extract bounded evidence lists from a risk payload.
 */
export function marketplaceRiskEvidence(
  risk: MarketplaceRiskFindingLike | null | undefined,
): MarketplaceRiskEvidence {
  const payload = risk?.payload || {};
  const files = asStringArray(payload.files);
  const services = asServiceEvidence(payload.services);
  const routes = asRouteEvidence(payload.routes);
  const totalCount =
    typeof payload.totalCount === "number" && payload.totalCount > 0
      ? payload.totalCount
      : Math.max(files.length, services.length, routes.length);
  return {
    example: typeof payload.example === "string" ? payload.example.trim() : "",
    expectedField:
      typeof payload.expectedField === "string"
        ? payload.expectedField.trim()
        : "",
    expectedPath:
      typeof payload.expectedPath === "string"
        ? payload.expectedPath.trim()
        : "",
    files,
    routes,
    services,
    totalCount,
    truncated: payload.truncated === true,
  };
}

export function marketplaceRiskHasEvidence(
  evidence: MarketplaceRiskEvidence | null | undefined,
): boolean {
  if (!evidence) {
    return false;
  }
  return (
    evidence.files.length > 0 ||
    evidence.services.length > 0 ||
    evidence.routes.length > 0 ||
    !!evidence.expectedPath ||
    !!evidence.expectedField ||
    !!evidence.example
  );
}

export type MarketplaceRiskDispositionCounts = {
  infoOnly: number;
  needAttention: number;
  needFix: number;
  blocking: number;
  total: number;
};

/**
 * Count findings by disposition for summary chips.
 */
export function countMarketplaceRiskDispositions(
  items: readonly MarketplaceRiskFindingLike[] | null | undefined,
): MarketplaceRiskDispositionCounts {
  const counts: MarketplaceRiskDispositionCounts = {
    blocking: 0,
    infoOnly: 0,
    needAttention: 0,
    needFix: 0,
    total: 0,
  };
  if (!items) {
    return counts;
  }
  for (const item of items) {
    counts.total += 1;
    if (marketplaceRiskBlocking(item)) {
      counts.blocking += 1;
    }
    switch (marketplaceRiskDisposition(item)) {
      case "need_fix": {
        counts.needFix += 1;
        break;
      }
      case "info_only": {
        counts.infoOnly += 1;
        break;
      }
      default: {
        counts.needAttention += 1;
      }
    }
  }
  return counts;
}

/**
 * Filter findings by disposition chip value (`all` returns everything).
 */
export function filterMarketplaceRiskFindingsByDisposition<
  T extends MarketplaceRiskFindingLike,
>(
  items: readonly T[] | null | undefined,
  disposition: "all" | MarketplaceRiskDisposition,
): T[] {
  const sorted = sortMarketplaceRiskFindingsBySeverity(items);
  if (!disposition || disposition === "all") {
    return sorted;
  }
  return sorted.filter(
    (item) => marketplaceRiskDisposition(item) === disposition,
  );
}
