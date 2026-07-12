// Shared marketplace workbench route helpers. Paths follow the top-level
// plugin-marketplace directory menu (above the host extension center).

const MARKETPLACE_BASE = "/plugin-marketplace";
const DEFAULT_BACK_PATH = "/dashboard/analytics";

export function marketplaceMinePath() {
  return `${MARKETPLACE_BASE}/plugin-marketplace-mine`;
}

export function marketplaceAdminListPath() {
  return `${MARKETPLACE_BASE}/plugin-marketplace-admin-list`;
}

export function marketplaceReviewPath() {
  return `${MARKETPLACE_BASE}/plugin-marketplace-review`;
}

export type MarketplaceDetailSource = "admin-list" | "mine" | "review";

export function marketplaceDetailPath(
  pluginId: string,
  from?: MarketplaceDetailSource,
) {
  const path =
    from === "mine"
      ? marketplaceMinePath()
      : from === "admin-list"
        ? marketplaceAdminListPath()
        : from === "review"
          ? marketplaceReviewPath()
          : `${MARKETPLACE_BASE}/plugin-marketplace-detail`;
  return {
    path,
    query: from
      ? { from, pageKey: path, pluginId, view: "detail" }
      : { pageKey: path, pluginId },
  };
}

export function marketplaceBackPath(from?: string | null) {
  if (from === "mine") {
    return marketplaceMinePath();
  }
  if (from === "admin-list") {
    return marketplaceAdminListPath();
  }
  if (from === "review") {
    return marketplaceReviewPath();
  }
  return DEFAULT_BACK_PATH;
}
