-- 001: Add Plugin Marketplace
-- 001：新增插件市场

-- ============================================================
-- Purpose: Stores plugin marketplace publisher profiles and ownership anchors.
-- 用途：存储插件市场发布者资料和归属锚点。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_publisher (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "publisher_key"   VARCHAR(64) NOT NULL DEFAULT '',
    "name"            VARCHAR(128) NOT NULL DEFAULT '',
    "summary"         VARCHAR(512) NOT NULL DEFAULT '',
    "owner_user_id"   BIGINT NOT NULL DEFAULT 0,
    "owner_org_id"    BIGINT NOT NULL DEFAULT 0,
    "verified"        BOOL NOT NULL DEFAULT FALSE,
    "status"          VARCHAR(32) NOT NULL DEFAULT 'active',
    "homepage"        VARCHAR(512) NOT NULL DEFAULT '',
    "contact_email"   VARCHAR(128) NOT NULL DEFAULT '',
    "remark"          VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMPTZ,
    "updated_at"      TIMESTAMPTZ,
    "deleted_at"      TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_publisher IS 'Plugin marketplace publisher table';
COMMENT ON COLUMN plugin_marketplace_publisher."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_publisher."publisher_key" IS 'Stable publisher key';
COMMENT ON COLUMN plugin_marketplace_publisher."name" IS 'Publisher display name';
COMMENT ON COLUMN plugin_marketplace_publisher."summary" IS 'Publisher summary';
COMMENT ON COLUMN plugin_marketplace_publisher."owner_user_id" IS 'Owning user ID';
COMMENT ON COLUMN plugin_marketplace_publisher."owner_org_id" IS 'Owning organization ID, 0 means none';
COMMENT ON COLUMN plugin_marketplace_publisher."verified" IS 'Whether the publisher has been verified';
COMMENT ON COLUMN plugin_marketplace_publisher."status" IS 'Publisher status: active/suspended';
COMMENT ON COLUMN plugin_marketplace_publisher."homepage" IS 'Publisher homepage URL';
COMMENT ON COLUMN plugin_marketplace_publisher."contact_email" IS 'Publisher contact email';
COMMENT ON COLUMN plugin_marketplace_publisher."remark" IS 'Remark';
COMMENT ON COLUMN plugin_marketplace_publisher."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_publisher."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_publisher."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_publisher_key ON plugin_marketplace_publisher ("publisher_key");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_publisher_owner ON plugin_marketplace_publisher ("owner_user_id", "status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_publisher_status ON plugin_marketplace_publisher ("status", "updated_at");

-- ============================================================
-- Purpose: Stores marketplace plugin identity records and plugin ID ownership.
-- 用途：存储市场插件身份记录和插件 ID 归属。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_plugin (
    "id"                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "publisher_id"      INT NOT NULL DEFAULT 0,
    "plugin_id"         VARCHAR(64) NOT NULL DEFAULT '',
    "name"              VARCHAR(128) NOT NULL DEFAULT '',
    "summary"           VARCHAR(512) NOT NULL DEFAULT '',
    "description"       TEXT,
    "plugin_type"       VARCHAR(32) NOT NULL DEFAULT 'source',
    "market_status"     VARCHAR(32) NOT NULL DEFAULT 'draft',
    "visibility"        VARCHAR(32) NOT NULL DEFAULT 'public',
    "latest_release_id" INT NOT NULL DEFAULT 0,
    "latest_version"    VARCHAR(32) NOT NULL DEFAULT '',
    "icon"              VARCHAR(512) NOT NULL DEFAULT '',
    "homepage"          VARCHAR(512) NOT NULL DEFAULT '',
    "repository"        VARCHAR(512) NOT NULL DEFAULT '',
    "license"           VARCHAR(64) NOT NULL DEFAULT '',
    "download_count"    BIGINT NOT NULL DEFAULT 0,
    "published_at"      TIMESTAMPTZ,
    "created_at"        TIMESTAMPTZ,
    "updated_at"        TIMESTAMPTZ,
    "deleted_at"        TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_plugin IS 'Marketplace plugin identity and ownership table';
COMMENT ON COLUMN plugin_marketplace_plugin."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_plugin."publisher_id" IS 'Owning publisher ID';
COMMENT ON COLUMN plugin_marketplace_plugin."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_plugin."name" IS 'Plugin display name';
COMMENT ON COLUMN plugin_marketplace_plugin."summary" IS 'Short marketplace summary';
COMMENT ON COLUMN plugin_marketplace_plugin."description" IS 'Long marketplace description';
COMMENT ON COLUMN plugin_marketplace_plugin."plugin_type" IS 'Plugin type: source/dynamic';
COMMENT ON COLUMN plugin_marketplace_plugin."market_status" IS 'Marketplace status: draft/published/delisted/deprecated';
COMMENT ON COLUMN plugin_marketplace_plugin."visibility" IS 'Visibility policy: public/private/reserved';
COMMENT ON COLUMN plugin_marketplace_plugin."latest_release_id" IS 'Latest published release ID';
COMMENT ON COLUMN plugin_marketplace_plugin."latest_version" IS 'Latest published version';
COMMENT ON COLUMN plugin_marketplace_plugin."icon" IS 'Marketplace icon path or URL';
COMMENT ON COLUMN plugin_marketplace_plugin."homepage" IS 'Plugin homepage URL';
COMMENT ON COLUMN plugin_marketplace_plugin."repository" IS 'Plugin source repository URL';
COMMENT ON COLUMN plugin_marketplace_plugin."license" IS 'Plugin license identifier';
COMMENT ON COLUMN plugin_marketplace_plugin."download_count" IS 'Aggregated download count snapshot';
COMMENT ON COLUMN plugin_marketplace_plugin."published_at" IS 'First published time';
COMMENT ON COLUMN plugin_marketplace_plugin."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_plugin."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_plugin."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_plugin_id ON plugin_marketplace_plugin ("plugin_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_publisher ON plugin_marketplace_plugin ("publisher_id", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_status_updated ON plugin_marketplace_plugin ("market_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_type_status ON plugin_marketplace_plugin ("plugin_type", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_visibility ON plugin_marketplace_plugin ("visibility", "market_status");

-- ============================================================
-- Purpose: Stores marketplace release records, review state, immutable version identity, and audit summaries.
-- 用途：存储市场发布版本、审核状态、不可变版本身份和审核摘要。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_release (
    "id"                   INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_record_id"     INT NOT NULL DEFAULT 0,
    "publisher_id"         INT NOT NULL DEFAULT 0,
    "plugin_id"            VARCHAR(64) NOT NULL DEFAULT '',
    "release_version"      VARCHAR(32) NOT NULL DEFAULT '',
    "plugin_type"          VARCHAR(32) NOT NULL DEFAULT 'source',
    "release_status"       VARCHAR(32) NOT NULL DEFAULT 'draft',
    "review_status"        VARCHAR(32) NOT NULL DEFAULT 'draft',
    "visibility"           VARCHAR(32) NOT NULL DEFAULT 'public',
    "min_host_version"     VARCHAR(32) NOT NULL DEFAULT '',
    "max_host_version"     VARCHAR(32) NOT NULL DEFAULT '',
    "manifest_snapshot"    JSONB NOT NULL DEFAULT '{}'::JSONB,
    "dependency_summary"   JSONB NOT NULL DEFAULT '[]'::JSONB,
    "host_service_summary" JSONB NOT NULL DEFAULT '[]'::JSONB,
    "route_summary"        JSONB NOT NULL DEFAULT '[]'::JSONB,
    "sql_summary"          JSONB NOT NULL DEFAULT '[]'::JSONB,
    "i18n_summary"         JSONB NOT NULL DEFAULT '[]'::JSONB,
    "docs_summary"         JSONB NOT NULL DEFAULT '[]'::JSONB,
    "risk_summary"         JSONB NOT NULL DEFAULT '{}'::JSONB,
    "review_message"       VARCHAR(1024) NOT NULL DEFAULT '',
    "submitted_at"         TIMESTAMPTZ,
    "reviewed_at"          TIMESTAMPTZ,
    "published_at"         TIMESTAMPTZ,
    "created_at"           TIMESTAMPTZ,
    "updated_at"           TIMESTAMPTZ,
    "deleted_at"           TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_release IS 'Marketplace plugin release table';
COMMENT ON COLUMN plugin_marketplace_release."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_release."plugin_record_id" IS 'Owning marketplace plugin record ID';
COMMENT ON COLUMN plugin_marketplace_release."publisher_id" IS 'Owning publisher ID';
COMMENT ON COLUMN plugin_marketplace_release."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_release."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_release."plugin_type" IS 'Plugin type: source/dynamic';
COMMENT ON COLUMN plugin_marketplace_release."release_status" IS 'Release status: draft/published/delisted/deprecated';
COMMENT ON COLUMN plugin_marketplace_release."review_status" IS 'Review status: draft/submitted/reviewing/approved/rejected';
COMMENT ON COLUMN plugin_marketplace_release."visibility" IS 'Release visibility policy';
COMMENT ON COLUMN plugin_marketplace_release."min_host_version" IS 'Minimum compatible LinaPro host version';
COMMENT ON COLUMN plugin_marketplace_release."max_host_version" IS 'Maximum compatible LinaPro host version';
COMMENT ON COLUMN plugin_marketplace_release."manifest_snapshot" IS 'Parsed plugin.yaml snapshot';
COMMENT ON COLUMN plugin_marketplace_release."dependency_summary" IS 'Dependency scan summary';
COMMENT ON COLUMN plugin_marketplace_release."host_service_summary" IS 'Host service scan summary';
COMMENT ON COLUMN plugin_marketplace_release."route_summary" IS 'Route scan summary';
COMMENT ON COLUMN plugin_marketplace_release."sql_summary" IS 'SQL resource scan summary';
COMMENT ON COLUMN plugin_marketplace_release."i18n_summary" IS 'i18n resource scan summary';
COMMENT ON COLUMN plugin_marketplace_release."docs_summary" IS 'Marketplace document scan summary';
COMMENT ON COLUMN plugin_marketplace_release."risk_summary" IS 'Aggregated review risk summary';
COMMENT ON COLUMN plugin_marketplace_release."review_message" IS 'Latest review message';
COMMENT ON COLUMN plugin_marketplace_release."submitted_at" IS 'Review submission time';
COMMENT ON COLUMN plugin_marketplace_release."reviewed_at" IS 'Review completion time';
COMMENT ON COLUMN plugin_marketplace_release."published_at" IS 'Publish time';
COMMENT ON COLUMN plugin_marketplace_release."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_release."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_release."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_release_plugin_version ON plugin_marketplace_release ("plugin_id", "release_version");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_plugin_status ON plugin_marketplace_release ("plugin_record_id", "release_status", "review_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_publisher_status ON plugin_marketplace_release ("publisher_id", "review_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_visibility ON plugin_marketplace_release ("visibility", "release_status");

-- ============================================================
-- Purpose: Stores uploaded marketplace artifacts and checksum metadata.
-- 用途：存储上传到市场的产物和校验和元数据。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_artifact (
    "id"                 INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "release_id"         INT NOT NULL DEFAULT 0,
    "plugin_id"          VARCHAR(64) NOT NULL DEFAULT '',
    "release_version"    VARCHAR(32) NOT NULL DEFAULT '',
    "artifact_type"      VARCHAR(32) NOT NULL DEFAULT '',
    "storage_key"        VARCHAR(512) NOT NULL DEFAULT '',
    "file_name"          VARCHAR(255) NOT NULL DEFAULT '',
    "content_type"       VARCHAR(128) NOT NULL DEFAULT '',
    "size_bytes"         BIGINT NOT NULL DEFAULT 0,
    "sha256"             VARCHAR(64) NOT NULL DEFAULT '',
    "manifest_sha256"    VARCHAR(64) NOT NULL DEFAULT '',
    "wasm_sha256"        VARCHAR(64) NOT NULL DEFAULT '',
    "created_at"         TIMESTAMPTZ,
    "updated_at"         TIMESTAMPTZ,
    "deleted_at"         TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_artifact IS 'Marketplace artifact table';
COMMENT ON COLUMN plugin_marketplace_artifact."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_artifact."release_id" IS 'Owning release ID';
COMMENT ON COLUMN plugin_marketplace_artifact."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_artifact."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_artifact."artifact_type" IS 'Artifact type: source_zip/dynamic_zip/plugin_wasm';
COMMENT ON COLUMN plugin_marketplace_artifact."storage_key" IS 'Storage object key or managed file key';
COMMENT ON COLUMN plugin_marketplace_artifact."file_name" IS 'Original artifact file name';
COMMENT ON COLUMN plugin_marketplace_artifact."content_type" IS 'Artifact content type';
COMMENT ON COLUMN plugin_marketplace_artifact."size_bytes" IS 'Artifact size in bytes';
COMMENT ON COLUMN plugin_marketplace_artifact."sha256" IS 'Artifact SHA-256 checksum';
COMMENT ON COLUMN plugin_marketplace_artifact."manifest_sha256" IS 'Root manifest SHA-256 checksum';
COMMENT ON COLUMN plugin_marketplace_artifact."wasm_sha256" IS 'Extracted plugin.wasm SHA-256 checksum';
COMMENT ON COLUMN plugin_marketplace_artifact."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_artifact."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_artifact."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_artifact_release_type ON plugin_marketplace_artifact ("release_id", "artifact_type");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_artifact_plugin ON plugin_marketplace_artifact ("plugin_id", "release_version");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_artifact_sha256 ON plugin_marketplace_artifact ("sha256");

-- ============================================================
-- Purpose: Stores marketplace document index entries by release and locale.
-- 用途：按版本和语言存储市场文档索引。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_doc (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "release_id"      INT NOT NULL DEFAULT 0,
    "plugin_id"       VARCHAR(64) NOT NULL DEFAULT '',
    "release_version" VARCHAR(32) NOT NULL DEFAULT '',
    "locale"          VARCHAR(32) NOT NULL DEFAULT '',
    "doc_path"        VARCHAR(255) NOT NULL DEFAULT '',
    "source_kind"     VARCHAR(32) NOT NULL DEFAULT '',
    "title"           VARCHAR(255) NOT NULL DEFAULT '',
    "summary"         VARCHAR(1024) NOT NULL DEFAULT '',
    "content_hash"    VARCHAR(64) NOT NULL DEFAULT '',
    "search_text"     TEXT,
    "created_at"      TIMESTAMPTZ,
    "updated_at"      TIMESTAMPTZ,
    "deleted_at"      TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_doc IS 'Marketplace document index table';
COMMENT ON COLUMN plugin_marketplace_doc."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_doc."release_id" IS 'Owning release ID';
COMMENT ON COLUMN plugin_marketplace_doc."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_doc."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_doc."locale" IS 'Document locale';
COMMENT ON COLUMN plugin_marketplace_doc."doc_path" IS 'Document path inside manifest/docs or README fallback';
COMMENT ON COLUMN plugin_marketplace_doc."source_kind" IS 'Document source kind: manifest_docs/readme';
COMMENT ON COLUMN plugin_marketplace_doc."title" IS 'Document title';
COMMENT ON COLUMN plugin_marketplace_doc."summary" IS 'Document search summary';
COMMENT ON COLUMN plugin_marketplace_doc."content_hash" IS 'Document content hash';
COMMENT ON COLUMN plugin_marketplace_doc."search_text" IS 'Plain text used for search indexing';
COMMENT ON COLUMN plugin_marketplace_doc."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_doc."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_doc."deleted_at" IS 'Deletion time';

DROP INDEX IF EXISTS uk_plugin_marketplace_doc_release_locale_path;
CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_doc_release_locale_path ON plugin_marketplace_doc ("release_id", "locale", "doc_path") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_doc_plugin_locale ON plugin_marketplace_doc ("plugin_id", "release_version", "locale");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_doc_hash ON plugin_marketplace_doc ("release_id", "content_hash");

-- ============================================================
-- Purpose: Stores review risk findings for release detail and review screens.
-- 用途：存储用于版本详情和审核页面的风险发现。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_risk (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "release_id"      INT NOT NULL DEFAULT 0,
    "plugin_id"       VARCHAR(64) NOT NULL DEFAULT '',
    "release_version" VARCHAR(32) NOT NULL DEFAULT '',
    "risk_type"       VARCHAR(64) NOT NULL DEFAULT '',
    "severity"        VARCHAR(32) NOT NULL DEFAULT 'info',
    "source"          VARCHAR(128) NOT NULL DEFAULT '',
    "summary"         VARCHAR(1024) NOT NULL DEFAULT '',
    "payload"         JSONB NOT NULL DEFAULT '{}'::JSONB,
    "created_at"      TIMESTAMPTZ,
    "updated_at"      TIMESTAMPTZ,
    "deleted_at"      TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_risk IS 'Marketplace release risk finding table';
COMMENT ON COLUMN plugin_marketplace_risk."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_risk."release_id" IS 'Owning release ID';
COMMENT ON COLUMN plugin_marketplace_risk."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_risk."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_risk."risk_type" IS 'Risk type: host_service/dynamic_route/menu_permission/external_network/data_table/install_sql/uninstall_sql/mock_sql/dependency/multi_tenant/docs';
COMMENT ON COLUMN plugin_marketplace_risk."severity" IS 'Risk severity: info/warning/high';
COMMENT ON COLUMN plugin_marketplace_risk."source" IS 'Scanner or resource source';
COMMENT ON COLUMN plugin_marketplace_risk."summary" IS 'Human-readable risk summary';
COMMENT ON COLUMN plugin_marketplace_risk."payload" IS 'Structured scanner payload';
COMMENT ON COLUMN plugin_marketplace_risk."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_risk."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_risk."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_risk_release ON plugin_marketplace_risk ("release_id", "severity", "risk_type");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_risk_plugin ON plugin_marketplace_risk ("plugin_id", "release_version");

-- ============================================================
-- Purpose: Stores marketplace category and tag definitions.
-- 用途：存储市场分类和标签定义。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_tag (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tag_code"    VARCHAR(64) NOT NULL DEFAULT '',
    "name"        VARCHAR(128) NOT NULL DEFAULT '',
    "tag_type"    VARCHAR(32) NOT NULL DEFAULT 'category',
    "sort"        INT NOT NULL DEFAULT 0,
    "status"      SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMPTZ,
    "updated_at"  TIMESTAMPTZ,
    "deleted_at"  TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_tag IS 'Marketplace category and tag table';
COMMENT ON COLUMN plugin_marketplace_tag."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_tag."tag_code" IS 'Stable tag code';
COMMENT ON COLUMN plugin_marketplace_tag."name" IS 'Tag display name';
COMMENT ON COLUMN plugin_marketplace_tag."tag_type" IS 'Tag type: category/tag';
COMMENT ON COLUMN plugin_marketplace_tag."sort" IS 'Display order';
COMMENT ON COLUMN plugin_marketplace_tag."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_marketplace_tag."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_tag."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_tag."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_tag_code ON plugin_marketplace_tag ("tag_code");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_tag_type_status ON plugin_marketplace_tag ("tag_type", "status", "sort");

CREATE TABLE IF NOT EXISTS plugin_marketplace_plugin_tag (
    "id"               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_record_id" INT NOT NULL DEFAULT 0,
    "plugin_id"        VARCHAR(64) NOT NULL DEFAULT '',
    "tag_code"         VARCHAR(64) NOT NULL DEFAULT '',
    "created_at"       TIMESTAMPTZ,
    "updated_at"       TIMESTAMPTZ,
    "deleted_at"       TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_plugin_tag IS 'Marketplace plugin tag relation table';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."plugin_record_id" IS 'Owning marketplace plugin record ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."tag_code" IS 'Stable tag code';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_plugin_tag ON plugin_marketplace_plugin_tag ("plugin_record_id", "tag_code");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_tag_code ON plugin_marketplace_plugin_tag ("tag_code", "plugin_record_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_tag_plugin ON plugin_marketplace_plugin_tag ("plugin_id", "tag_code");

-- ============================================================
-- Purpose: Stores visibility grants for private and future licensed marketplace plugins.
-- 用途：存储私有插件和未来授权插件的可见性授权。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_visibility_grant (
    "id"               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_record_id" INT NOT NULL DEFAULT 0,
    "plugin_id"        VARCHAR(64) NOT NULL DEFAULT '',
    "scope_type"       VARCHAR(32) NOT NULL DEFAULT '',
    "scope_id"         VARCHAR(128) NOT NULL DEFAULT '',
    "permission"       VARCHAR(32) NOT NULL DEFAULT 'view',
    "status"           SMALLINT NOT NULL DEFAULT 1,
    "expires_at"       TIMESTAMPTZ,
    "created_at"       TIMESTAMPTZ,
    "updated_at"       TIMESTAMPTZ,
    "deleted_at"       TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_visibility_grant IS 'Marketplace visibility grant table';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."plugin_record_id" IS 'Owning marketplace plugin record ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."scope_type" IS 'Visibility scope type: public/tenant/org/user/reserved_license';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."scope_id" IS 'Scope identifier, empty for public scope';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."permission" IS 'Permission covered by the grant: view/download';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."expires_at" IS 'Grant expiration time';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_visibility_grant ON plugin_marketplace_visibility_grant ("plugin_record_id", "scope_type", "scope_id", "permission");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_visibility_lookup ON plugin_marketplace_visibility_grant ("scope_type", "scope_id", "permission", "status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_visibility_plugin ON plugin_marketplace_visibility_grant ("plugin_id", "permission", "status");

-- ============================================================
-- Purpose: Stores short-lived download sessions and authorization snapshots.
-- 用途：存储短期下载会话和授权快照。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_download_session (
    "id"                         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "session_id"                 VARCHAR(64) NOT NULL DEFAULT '',
    "release_id"                 INT NOT NULL DEFAULT 0,
    "artifact_id"                INT NOT NULL DEFAULT 0,
    "plugin_id"                  VARCHAR(64) NOT NULL DEFAULT '',
    "release_version"            VARCHAR(32) NOT NULL DEFAULT '',
    "requester_user_id"          BIGINT NOT NULL DEFAULT 0,
    "status"                     VARCHAR(32) NOT NULL DEFAULT 'active',
    "artifact_type"              VARCHAR(32) NOT NULL DEFAULT '',
    "artifact_size_bytes"        BIGINT NOT NULL DEFAULT 0,
    "sha256"                     VARCHAR(64) NOT NULL DEFAULT '',
    "authorization_snapshot"     JSONB NOT NULL DEFAULT '{}'::JSONB,
    "expires_at"                 TIMESTAMPTZ,
    "consumed_at"                TIMESTAMPTZ,
    "created_at"                 TIMESTAMPTZ,
    "updated_at"                 TIMESTAMPTZ,
    "deleted_at"                 TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_download_session IS 'Marketplace download session table';
COMMENT ON COLUMN plugin_marketplace_download_session."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_download_session."session_id" IS 'Opaque download session ID';
COMMENT ON COLUMN plugin_marketplace_download_session."release_id" IS 'Owning release ID';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_id" IS 'Owning artifact ID';
COMMENT ON COLUMN plugin_marketplace_download_session."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_download_session."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_download_session."requester_user_id" IS 'Requester user ID';
COMMENT ON COLUMN plugin_marketplace_download_session."status" IS 'Session status: active/expired/consumed/revoked';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_type" IS 'Artifact type bound to the session';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_size_bytes" IS 'Artifact size in bytes';
COMMENT ON COLUMN plugin_marketplace_download_session."sha256" IS 'Artifact SHA-256 checksum returned to the client';
COMMENT ON COLUMN plugin_marketplace_download_session."authorization_snapshot" IS 'Authorization decision snapshot captured at session creation';
COMMENT ON COLUMN plugin_marketplace_download_session."expires_at" IS 'Session expiration time';
COMMENT ON COLUMN plugin_marketplace_download_session."consumed_at" IS 'First successful download time';
COMMENT ON COLUMN plugin_marketplace_download_session."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_download_session."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_download_session."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_download_session_id ON plugin_marketplace_download_session ("session_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_expire ON plugin_marketplace_download_session ("status", "expires_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_requester ON plugin_marketplace_download_session ("requester_user_id", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_release ON plugin_marketplace_download_session ("release_id", "status");

-- ============================================================
-- Purpose: Stores download events that are asynchronously aggregated into read-model counters.
-- 用途：存储下载事件，并异步聚合到读模型计数字段。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_download_event (
    "id"                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "session_id"        VARCHAR(64) NOT NULL DEFAULT '',
    "release_id"        INT NOT NULL DEFAULT 0,
    "artifact_id"       INT NOT NULL DEFAULT 0,
    "plugin_id"         VARCHAR(64) NOT NULL DEFAULT '',
    "release_version"   VARCHAR(32) NOT NULL DEFAULT '',
    "requester_user_id" BIGINT NOT NULL DEFAULT 0,
    "event_type"        VARCHAR(32) NOT NULL DEFAULT '',
    "client_ip_hash"    VARCHAR(128) NOT NULL DEFAULT '',
    "user_agent_hash"   VARCHAR(128) NOT NULL DEFAULT '',
    "created_at"        TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_download_event IS 'Marketplace download event table';
COMMENT ON COLUMN plugin_marketplace_download_event."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_download_event."session_id" IS 'Opaque download session ID';
COMMENT ON COLUMN plugin_marketplace_download_event."release_id" IS 'Owning release ID';
COMMENT ON COLUMN plugin_marketplace_download_event."artifact_id" IS 'Owning artifact ID';
COMMENT ON COLUMN plugin_marketplace_download_event."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_download_event."release_version" IS 'Plugin release version';
COMMENT ON COLUMN plugin_marketplace_download_event."requester_user_id" IS 'Requester user ID';
COMMENT ON COLUMN plugin_marketplace_download_event."event_type" IS 'Download event type: created/started/completed/failed';
COMMENT ON COLUMN plugin_marketplace_download_event."client_ip_hash" IS 'Hashed client IP for statistics';
COMMENT ON COLUMN plugin_marketplace_download_event."user_agent_hash" IS 'Hashed user agent for statistics';
COMMENT ON COLUMN plugin_marketplace_download_event."created_at" IS 'Creation time';

CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_release ON plugin_marketplace_download_event ("release_id", "event_type", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_plugin ON plugin_marketplace_download_event ("plugin_id", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_session ON plugin_marketplace_download_event ("session_id", "event_type");

-- ============================================================
-- Purpose: Stores the marketplace list projection used by paginated catalog APIs.
-- 用途：存储市场分页目录接口使用的列表投影。
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_plugin_read_model (
    "id"                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_record_id"  INT NOT NULL DEFAULT 0,
    "publisher_id"      INT NOT NULL DEFAULT 0,
    "publisher_name"    VARCHAR(128) NOT NULL DEFAULT '',
    "publisher_verified" BOOL NOT NULL DEFAULT FALSE,
    "plugin_id"         VARCHAR(64) NOT NULL DEFAULT '',
    "name"              VARCHAR(128) NOT NULL DEFAULT '',
    "summary"           VARCHAR(512) NOT NULL DEFAULT '',
    "plugin_type"       VARCHAR(32) NOT NULL DEFAULT 'source',
    "market_status"     VARCHAR(32) NOT NULL DEFAULT 'draft',
    "visibility"        VARCHAR(32) NOT NULL DEFAULT 'public',
    "latest_release_id" INT NOT NULL DEFAULT 0,
    "latest_version"    VARCHAR(32) NOT NULL DEFAULT '',
    "min_host_version"  VARCHAR(32) NOT NULL DEFAULT '',
    "max_host_version"  VARCHAR(32) NOT NULL DEFAULT '',
    "primary_tag"       VARCHAR(64) NOT NULL DEFAULT '',
    "tag_codes"         JSONB NOT NULL DEFAULT '[]'::JSONB,
    "risk_counts"       JSONB NOT NULL DEFAULT '{}'::JSONB,
    "download_count"    BIGINT NOT NULL DEFAULT 0,
    "published_at"      TIMESTAMPTZ,
    "search_text"       TEXT,
    "created_at"        TIMESTAMPTZ,
    "updated_at"        TIMESTAMPTZ,
    "deleted_at"        TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_plugin_read_model IS 'Marketplace plugin list read model table';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."id" IS 'Primary key ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_record_id" IS 'Owning marketplace plugin record ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_id" IS 'Owning publisher ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_name" IS 'Publisher display name snapshot';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_verified" IS 'Publisher verification snapshot';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_id" IS 'Stable plugin ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."name" IS 'Plugin display name';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."summary" IS 'Short marketplace summary';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_type" IS 'Plugin type: source/dynamic';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."market_status" IS 'Marketplace status';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."visibility" IS 'Visibility policy';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."latest_release_id" IS 'Latest published release ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."latest_version" IS 'Latest published version';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."min_host_version" IS 'Minimum compatible LinaPro host version';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."max_host_version" IS 'Maximum compatible LinaPro host version';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."primary_tag" IS 'Primary category tag code';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."tag_codes" IS 'Tag code snapshot for display';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."risk_counts" IS 'Risk count snapshot grouped by severity';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."download_count" IS 'Aggregated download count snapshot';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."published_at" IS 'Latest publish time';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."search_text" IS 'Plain text projection used for catalog search';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_read_model_plugin ON plugin_marketplace_plugin_read_model ("plugin_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_status_updated ON plugin_marketplace_plugin_read_model ("market_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_type_status ON plugin_marketplace_plugin_read_model ("plugin_type", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_primary_tag ON plugin_marketplace_plugin_read_model ("primary_tag", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_visibility ON plugin_marketplace_plugin_read_model ("visibility", "market_status");

-- ============================================================
-- Dictionary seed data
-- 字典初始化数据
-- ============================================================
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace plugin type', 'plugin_marketplace_plugin_type', 1, 1, 'Plugin marketplace plugin type list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace status', 'plugin_marketplace_status', 1, 1, 'Plugin marketplace status list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace review status', 'plugin_marketplace_review_status', 1, 1, 'Plugin marketplace review status list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace artifact type', 'plugin_marketplace_artifact_type', 1, 1, 'Plugin marketplace artifact type list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace risk type', 'plugin_marketplace_risk_type', 1, 1, 'Plugin marketplace risk type list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace visibility', 'plugin_marketplace_visibility', 1, 1, 'Plugin marketplace visibility list', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, 'Plugin marketplace download session status', 'plugin_marketplace_download_session_status', 1, 1, 'Plugin marketplace download session status list', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_plugin_type', 'Source plugin', 'source', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_plugin_type', 'Dynamic plugin', 'dynamic', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', 'Draft', 'draft', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', 'Published', 'published', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', 'Delisted', 'delisted', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', 'Deprecated', 'deprecated', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', 'Draft', 'draft', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', 'Submitted', 'submitted', 2, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', 'Reviewing', 'reviewing', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', 'Approved', 'approved', 4, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', 'Rejected', 'rejected', 5, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', 'Source ZIP', 'source_zip', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', 'Dynamic ZIP', 'dynamic_zip', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', 'Plugin WASM', 'plugin_wasm', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Host service', 'host_service', 1, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Dynamic route', 'dynamic_route', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Menu permission', 'menu_permission', 3, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'External network', 'external_network', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Data table', 'data_table', 5, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Install SQL', 'install_sql', 6, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Uninstall SQL', 'uninstall_sql', 7, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Mock SQL', 'mock_sql', 8, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Dependency', 'dependency', 9, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Multi tenant', 'multi_tenant', 10, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Documentation', 'docs', 11, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', 'Public', 'public', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', 'Private', 'private', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', 'Reserved license', 'reserved', 3, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', 'Active', 'active', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', 'Expired', 'expired', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', 'Consumed', 'consumed', 3, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', 'Revoked', 'revoked', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
