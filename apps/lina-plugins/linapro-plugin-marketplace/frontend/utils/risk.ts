/**
 * Marketplace risk finding display helpers.
 *
 * Scanner rows persist English source text in `summary` and a stable diagnostic
 * code in `payload.code`. UI locales must render the code through plugin i18n
 * keys so Chinese (and other locales) never fall back to the stored English
 * source when a translation exists.
 *
 * Risk list presentation also sorts findings by severity so high-impact items
 * appear before warnings and informational notes.
 */

export type MarketplaceRiskSeverityRank = "high" | "info" | "warning" | string;

export type MarketplaceRiskFindingLike = {
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

/**
 * Full runtime i18n keys for every known scanner finding code. Kept as a static
 * string list so i18n coverage checks can discover keys without executing Vue.
 */
export const MARKETPLACE_RISK_FINDING_I18N_KEYS =
  MARKETPLACE_RISK_FINDING_CODES.map(
    (code) =>
      `plugin.linapro-plugin-marketplace.detail.riskFinding.${code}` as const,
  );

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
 * Stable sort of risk findings by severity (high first). Same-severity items
 * keep their original relative order.
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
      const rankDiff =
        marketplaceRiskSeverityRank(left.item.severity) -
        marketplaceRiskSeverityRank(right.item.severity);
      if (rankDiff !== 0) {
        return rankDiff;
      }
      return left.index - right.index;
    })
    .map((entry) => entry.item);
}

/**
 * Resolve the stable scanner diagnostic code stored on a risk finding payload.
 * Returns an empty string when the payload is missing or not a non-empty string.
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
 * Build the runtime i18n key for one scanner diagnostic code.
 */
export function marketplaceRiskFindingMessageKey(code: string): string {
  const normalized = code.trim();
  if (!normalized) {
    return "";
  }
  return `${riskFindingKeyPrefix}${normalized}`;
}

/**
 * Localize one risk finding summary for the current UI locale.
 *
 * Prefer `payload.code` → `detail.riskFinding.<code>` so stored English source
 * text is only used as a fallback for unknown/legacy rows without a code or
 * without a translation entry.
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
  if (!key) {
    return fallback;
  }
  if (typeof te === "function") {
    if (!te(key)) {
      return fallback;
    }
    return t(key);
  }
  const translated = t(key);
  if (!translated || translated === key) {
    return fallback;
  }
  return translated;
}
