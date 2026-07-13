-- ============================================================
-- Change: simplify-plugin-marketplace-distribution
-- Purpose: Add publish source kind, Git repository metadata, sync state,
--          optional encrypted credential storage, and release source ref.
-- 用途：增加发布来源类型、Git 仓库元数据、同步状态、可选加密凭证
--       以及版本来源引用字段。
-- ============================================================

-- plugin_marketplace_plugin: publish source and Git sync metadata
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "source_kind" VARCHAR(32) NOT NULL DEFAULT 'upload';
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "repo_url" VARCHAR(512) NOT NULL DEFAULT '';
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "repo_provider" VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "credential_ref" VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "last_sync_at" TIMESTAMPTZ;
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "last_sync_status" VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE plugin_marketplace_plugin ADD COLUMN IF NOT EXISTS "last_sync_message" VARCHAR(1024) NOT NULL DEFAULT '';

COMMENT ON COLUMN plugin_marketplace_plugin."source_kind" IS 'Publish source kind: git/upload';
COMMENT ON COLUMN plugin_marketplace_plugin."repo_url" IS 'Git repository URL when source_kind is git';
COMMENT ON COLUMN plugin_marketplace_plugin."repo_provider" IS 'Git provider: github/gitee, empty for upload';
COMMENT ON COLUMN plugin_marketplace_plugin."credential_ref" IS 'Opaque credential reference for private Git access, empty when public';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_at" IS 'Last Git metadata discovery time';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_status" IS 'Last Git sync status: success/failed/auth_failed/partial, empty when never synced';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_message" IS 'Last Git sync diagnostic message without secrets';

UPDATE plugin_marketplace_plugin
SET "source_kind" = 'upload'
WHERE COALESCE(TRIM("source_kind"), '') = '';

CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_source_kind
    ON plugin_marketplace_plugin ("source_kind", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_repo_provider
    ON plugin_marketplace_plugin ("repo_provider", "source_kind")
    WHERE "source_kind" = 'git';

-- plugin_marketplace_release: git tag/ref for distribution.mode=git
ALTER TABLE plugin_marketplace_release ADD COLUMN IF NOT EXISTS "source_ref" VARCHAR(128) NOT NULL DEFAULT '';
COMMENT ON COLUMN plugin_marketplace_release."source_ref" IS 'Git tag or ref for git-sourced releases, empty for upload packages';

-- Secure credential store for private Git tokens (token ciphertext never returned by APIs)
CREATE TABLE IF NOT EXISTS plugin_marketplace_credential (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "credential_ref"  VARCHAR(64) NOT NULL DEFAULT '',
    "owner_user_id"   BIGINT NOT NULL DEFAULT 0,
    "provider"        VARCHAR(32) NOT NULL DEFAULT '',
    "cipher_text"     TEXT NOT NULL DEFAULT '',
    "created_at"      TIMESTAMPTZ,
    "updated_at"      TIMESTAMPTZ,
    "deleted_at"      TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_credential IS 'Marketplace encrypted Git credential table';
COMMENT ON COLUMN plugin_marketplace_credential."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_credential."credential_ref" IS 'Opaque credential reference stored on plugin records';
COMMENT ON COLUMN plugin_marketplace_credential."owner_user_id" IS 'Owning user ID of the credential';
COMMENT ON COLUMN plugin_marketplace_credential."provider" IS 'Git provider associated with the credential';
COMMENT ON COLUMN plugin_marketplace_credential."cipher_text" IS 'Encrypted token ciphertext; never returned by marketplace APIs';
COMMENT ON COLUMN plugin_marketplace_credential."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_credential."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_credential."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_credential_ref
    ON plugin_marketplace_credential ("credential_ref");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_credential_owner
    ON plugin_marketplace_credential ("owner_user_id", "provider");
