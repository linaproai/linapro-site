-- Supports the cross-publisher review queue filter and stable newest-first order.
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_review_queue
    ON plugin_marketplace_release (
        "review_status",
        "submitted_at" DESC,
        "updated_at" DESC,
        "id" DESC
    )
    WHERE "deleted_at" IS NULL;
