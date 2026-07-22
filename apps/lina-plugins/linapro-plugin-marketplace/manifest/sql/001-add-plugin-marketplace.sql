-- 001: 插件市场
-- 用途：安装插件市场业务表、索引与字典初始化数据（最终态，不考虑历史兼容迁移）。

-- ============================================================
-- 发布者资料与归属锚点
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

COMMENT ON TABLE plugin_marketplace_publisher IS '插件市场发布者表';
COMMENT ON COLUMN plugin_marketplace_publisher."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_publisher."publisher_key" IS '稳定发布者标识';
COMMENT ON COLUMN plugin_marketplace_publisher."name" IS '发布者展示名称';
COMMENT ON COLUMN plugin_marketplace_publisher."summary" IS '发布者简介';
COMMENT ON COLUMN plugin_marketplace_publisher."owner_user_id" IS '归属用户 ID';
COMMENT ON COLUMN plugin_marketplace_publisher."owner_org_id" IS '归属组织 ID，0 表示无';
COMMENT ON COLUMN plugin_marketplace_publisher."verified" IS '是否已认证';
COMMENT ON COLUMN plugin_marketplace_publisher."status" IS '发布者状态：active/suspended';
COMMENT ON COLUMN plugin_marketplace_publisher."homepage" IS '发布者主页 URL';
COMMENT ON COLUMN plugin_marketplace_publisher."contact_email" IS '发布者联系邮箱';
COMMENT ON COLUMN plugin_marketplace_publisher."remark" IS '备注';
COMMENT ON COLUMN plugin_marketplace_publisher."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_publisher."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_publisher."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_publisher_key ON plugin_marketplace_publisher ("publisher_key");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_publisher_owner ON plugin_marketplace_publisher ("owner_user_id", "status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_publisher_status ON plugin_marketplace_publisher ("status", "updated_at");

-- ============================================================
-- 市场插件身份、归属、发布来源与异步处理状态
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
    "process_status"    VARCHAR(32) NOT NULL DEFAULT 'pending_verify',
    "source_kind"       VARCHAR(32) NOT NULL DEFAULT 'upload',
    "repo_url"          VARCHAR(512) NOT NULL DEFAULT '',
    "repo_provider"     VARCHAR(32) NOT NULL DEFAULT '',
    "repo_path"         VARCHAR(512) NOT NULL DEFAULT '',
    "credential_ref"    VARCHAR(64) NOT NULL DEFAULT '',
    "last_sync_at"      TIMESTAMPTZ,
    "last_sync_status"  VARCHAR(32) NOT NULL DEFAULT '',
    "last_sync_message" VARCHAR(1024) NOT NULL DEFAULT '',
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

COMMENT ON TABLE plugin_marketplace_plugin IS '插件市场插件身份与归属表';
COMMENT ON COLUMN plugin_marketplace_plugin."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_plugin."publisher_id" IS '归属发布者 ID';
COMMENT ON COLUMN plugin_marketplace_plugin."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_plugin."name" IS '插件展示名称';
COMMENT ON COLUMN plugin_marketplace_plugin."summary" IS '市场短简介';
COMMENT ON COLUMN plugin_marketplace_plugin."description" IS '市场详细描述';
COMMENT ON COLUMN plugin_marketplace_plugin."plugin_type" IS '插件类型：source/dynamic';
COMMENT ON COLUMN plugin_marketplace_plugin."market_status" IS '市场状态：draft/published/delisted/deprecated';
COMMENT ON COLUMN plugin_marketplace_plugin."visibility" IS '可见性策略：public/private/reserved';
COMMENT ON COLUMN plugin_marketplace_plugin."process_status" IS '异步处理状态：pending_verify/pending_review/completed/failed';
COMMENT ON COLUMN plugin_marketplace_plugin."source_kind" IS '发布来源类型：git/upload';
COMMENT ON COLUMN plugin_marketplace_plugin."repo_url" IS 'Git 仓库 URL（source_kind 为 git 时）';
COMMENT ON COLUMN plugin_marketplace_plugin."repo_provider" IS 'Git 提供商：github/gitee，上传来源时为空';
COMMENT ON COLUMN plugin_marketplace_plugin."repo_path" IS '插件根目录相对仓库根路径；仓库根即插件根时为空';
COMMENT ON COLUMN plugin_marketplace_plugin."credential_ref" IS '私有 Git 凭证不透明引用，公开仓库为空';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_at" IS '最近一次 Git 元数据发现时间';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_status" IS '最近一次 Git 同步状态：success/failed/auth_failed/partial，从未同步时为空';
COMMENT ON COLUMN plugin_marketplace_plugin."last_sync_message" IS '最近一次 Git 同步诊断信息（不含密钥）';
COMMENT ON COLUMN plugin_marketplace_plugin."latest_release_id" IS '最新已发布版本 ID';
COMMENT ON COLUMN plugin_marketplace_plugin."latest_version" IS '最新已发布版本号';
COMMENT ON COLUMN plugin_marketplace_plugin."icon" IS '市场图标路径或 URL';
COMMENT ON COLUMN plugin_marketplace_plugin."homepage" IS '插件主页 URL';
COMMENT ON COLUMN plugin_marketplace_plugin."repository" IS '插件源码仓库 URL';
COMMENT ON COLUMN plugin_marketplace_plugin."license" IS '插件许可证标识';
COMMENT ON COLUMN plugin_marketplace_plugin."download_count" IS '聚合下载次数快照';
COMMENT ON COLUMN plugin_marketplace_plugin."published_at" IS '首次发布时间';
COMMENT ON COLUMN plugin_marketplace_plugin."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_plugin."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_plugin."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_plugin_id ON plugin_marketplace_plugin ("plugin_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_publisher ON plugin_marketplace_plugin ("publisher_id", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_status_updated ON plugin_marketplace_plugin ("market_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_type_status ON plugin_marketplace_plugin ("plugin_type", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_visibility ON plugin_marketplace_plugin ("visibility", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_source_kind ON plugin_marketplace_plugin ("source_kind", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_repo_provider ON plugin_marketplace_plugin ("repo_provider", "source_kind") WHERE "source_kind" = 'git';
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_process_status ON plugin_marketplace_plugin ("process_status", "updated_at");

-- ============================================================
-- 市场发布版本、审核状态、不可变版本身份与审核摘要
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
    "process_status"       VARCHAR(32) NOT NULL DEFAULT 'pending_verify',
    "visibility"           VARCHAR(32) NOT NULL DEFAULT 'public',
    "source_ref"           VARCHAR(128) NOT NULL DEFAULT '',
    "source_commit"        VARCHAR(64) NOT NULL DEFAULT '',
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

COMMENT ON TABLE plugin_marketplace_release IS '插件市场发布版本表';
COMMENT ON COLUMN plugin_marketplace_release."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_release."plugin_record_id" IS '归属市场插件记录 ID';
COMMENT ON COLUMN plugin_marketplace_release."publisher_id" IS '归属发布者 ID';
COMMENT ON COLUMN plugin_marketplace_release."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_release."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_release."plugin_type" IS '插件类型：source/dynamic';
COMMENT ON COLUMN plugin_marketplace_release."release_status" IS '版本状态：draft/published/delisted/deprecated';
COMMENT ON COLUMN plugin_marketplace_release."review_status" IS '审核状态：draft/submitted/reviewing/approved/rejected';
COMMENT ON COLUMN plugin_marketplace_release."process_status" IS '异步处理状态：pending_verify/pending_review/completed/failed';
COMMENT ON COLUMN plugin_marketplace_release."visibility" IS '版本可见性策略';
COMMENT ON COLUMN plugin_marketplace_release."source_ref" IS 'Git 来源版本的逻辑 tag/分支名（如 v1.0.0 或 main），上传包为空';
COMMENT ON COLUMN plugin_marketplace_release."source_commit" IS 'Git 发现时钉扎的完整 commit SHA，安装 distribution.ref 优先使用；上传包为空';
COMMENT ON COLUMN plugin_marketplace_release."min_host_version" IS '最低兼容 LinaPro 宿主版本';
COMMENT ON COLUMN plugin_marketplace_release."max_host_version" IS '最高兼容 LinaPro 宿主版本';
COMMENT ON COLUMN plugin_marketplace_release."manifest_snapshot" IS '解析后的 plugin.yaml 快照';
COMMENT ON COLUMN plugin_marketplace_release."dependency_summary" IS '依赖扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."host_service_summary" IS '宿主服务扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."route_summary" IS '路由扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."sql_summary" IS 'SQL 资源扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."i18n_summary" IS 'i18n 资源扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."docs_summary" IS '市场文档扫描摘要';
COMMENT ON COLUMN plugin_marketplace_release."risk_summary" IS '聚合审核风险摘要';
COMMENT ON COLUMN plugin_marketplace_release."review_message" IS '最近一次审核意见';
COMMENT ON COLUMN plugin_marketplace_release."submitted_at" IS '提交审核时间';
COMMENT ON COLUMN plugin_marketplace_release."reviewed_at" IS '审核完成时间';
COMMENT ON COLUMN plugin_marketplace_release."published_at" IS '发布时间';
COMMENT ON COLUMN plugin_marketplace_release."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_release."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_release."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_release_plugin_version ON plugin_marketplace_release ("plugin_id", "release_version");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_plugin_status ON plugin_marketplace_release ("plugin_record_id", "release_status", "review_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_publisher_status ON plugin_marketplace_release ("publisher_id", "review_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_visibility ON plugin_marketplace_release ("visibility", "release_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_process_status ON plugin_marketplace_release ("process_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_release_review_queue
    ON plugin_marketplace_release (
        "review_status",
        "submitted_at" DESC,
        "updated_at" DESC,
        "id" DESC
    )
    WHERE "deleted_at" IS NULL;

-- ============================================================
-- 市场产物与校验和元数据
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

COMMENT ON TABLE plugin_marketplace_artifact IS '插件市场产物表';
COMMENT ON COLUMN plugin_marketplace_artifact."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_artifact."release_id" IS '归属版本 ID';
COMMENT ON COLUMN plugin_marketplace_artifact."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_artifact."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_artifact."artifact_type" IS '产物类型：source_zip/dynamic_zip/plugin_wasm';
COMMENT ON COLUMN plugin_marketplace_artifact."storage_key" IS '存储对象键或托管文件键';
COMMENT ON COLUMN plugin_marketplace_artifact."file_name" IS '原始产物文件名';
COMMENT ON COLUMN plugin_marketplace_artifact."content_type" IS '产物内容类型';
COMMENT ON COLUMN plugin_marketplace_artifact."size_bytes" IS '产物大小（字节）';
COMMENT ON COLUMN plugin_marketplace_artifact."sha256" IS '产物 SHA-256 校验和';
COMMENT ON COLUMN plugin_marketplace_artifact."manifest_sha256" IS '根清单 SHA-256 校验和';
COMMENT ON COLUMN plugin_marketplace_artifact."wasm_sha256" IS '提取的 plugin.wasm SHA-256 校验和';
COMMENT ON COLUMN plugin_marketplace_artifact."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_artifact."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_artifact."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_artifact_release_type ON plugin_marketplace_artifact ("release_id", "artifact_type");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_artifact_plugin ON plugin_marketplace_artifact ("plugin_id", "release_version");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_artifact_sha256 ON plugin_marketplace_artifact ("sha256");

-- ============================================================
-- 私有 Git 凭证密文存储（API 永不返回密文）
-- ============================================================
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

COMMENT ON TABLE plugin_marketplace_credential IS '插件市场加密 Git 凭证表';
COMMENT ON COLUMN plugin_marketplace_credential."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_credential."credential_ref" IS '挂在插件记录上的不透明凭证引用';
COMMENT ON COLUMN plugin_marketplace_credential."owner_user_id" IS '凭证归属用户 ID';
COMMENT ON COLUMN plugin_marketplace_credential."provider" IS '凭证关联的 Git 提供商';
COMMENT ON COLUMN plugin_marketplace_credential."cipher_text" IS '加密后的 token 密文；市场 API 永不返回';
COMMENT ON COLUMN plugin_marketplace_credential."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_credential."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_credential."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_credential_ref ON plugin_marketplace_credential ("credential_ref");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_credential_owner ON plugin_marketplace_credential ("owner_user_id", "provider");

-- ============================================================
-- 版本展示元数据（名称/摘要多语言）；文档正文与图片不入库，权威在制品磁盘
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_display_i18n (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "release_id"      INT NOT NULL DEFAULT 0,
    "plugin_id"       VARCHAR(64) NOT NULL DEFAULT '',
    "release_version" VARCHAR(32) NOT NULL DEFAULT '',
    "locale"          VARCHAR(32) NOT NULL DEFAULT '',
    "name"            VARCHAR(128) NOT NULL DEFAULT '',
    "summary"         VARCHAR(512) NOT NULL DEFAULT '',
    "source"          VARCHAR(32) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMPTZ,
    "updated_at"      TIMESTAMPTZ,
    "deleted_at"      TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_display_i18n IS '插件市场版本展示元数据多语言表（仅名称与摘要）';
COMMENT ON COLUMN plugin_marketplace_display_i18n."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_display_i18n."release_id" IS '归属版本 ID';
COMMENT ON COLUMN plugin_marketplace_display_i18n."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_display_i18n."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_display_i18n."locale" IS '展示语言';
COMMENT ON COLUMN plugin_marketplace_display_i18n."name" IS '本地化展示名称';
COMMENT ON COLUMN plugin_marketplace_display_i18n."summary" IS '本地化列表摘要';
COMMENT ON COLUMN plugin_marketplace_display_i18n."source" IS '来源：package_i18n/plugin_yaml/publisher';
COMMENT ON COLUMN plugin_marketplace_display_i18n."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_display_i18n."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_display_i18n."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_display_i18n_release_locale
    ON plugin_marketplace_display_i18n ("release_id", "locale")
    WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_display_i18n_plugin_version
    ON plugin_marketplace_display_i18n ("plugin_id", "release_version", "locale");

-- ============================================================
-- 版本详情与审核页使用的风险发现
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

COMMENT ON TABLE plugin_marketplace_risk IS '插件市场版本风险发现表';
COMMENT ON COLUMN plugin_marketplace_risk."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_risk."release_id" IS '归属版本 ID';
COMMENT ON COLUMN plugin_marketplace_risk."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_risk."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_risk."risk_type" IS '风险类型：host_service/dynamic_route/menu_permission/external_network/data_table/install_sql/uninstall_sql/mock_sql/dependency/multi_tenant/docs';
COMMENT ON COLUMN plugin_marketplace_risk."severity" IS '风险级别：info/warning/high';
COMMENT ON COLUMN plugin_marketplace_risk."source" IS '扫描器或资源来源';
COMMENT ON COLUMN plugin_marketplace_risk."summary" IS '可读风险摘要';
COMMENT ON COLUMN plugin_marketplace_risk."payload" IS '结构化扫描载荷';
COMMENT ON COLUMN plugin_marketplace_risk."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_risk."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_risk."deleted_at" IS '删除时间';

CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_risk_release ON plugin_marketplace_risk ("release_id", "severity", "risk_type");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_risk_plugin ON plugin_marketplace_risk ("plugin_id", "release_version");

-- ============================================================
-- 市场分类与标签定义
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

COMMENT ON TABLE plugin_marketplace_tag IS '插件市场分类与标签表';
COMMENT ON COLUMN plugin_marketplace_tag."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_tag."tag_code" IS '稳定标签编码';
COMMENT ON COLUMN plugin_marketplace_tag."name" IS '标签展示名称';
COMMENT ON COLUMN plugin_marketplace_tag."tag_type" IS '标签类型：category/tag';
COMMENT ON COLUMN plugin_marketplace_tag."sort" IS '展示排序';
COMMENT ON COLUMN plugin_marketplace_tag."status" IS '状态：0=停用，1=启用';
COMMENT ON COLUMN plugin_marketplace_tag."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_tag."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_tag."deleted_at" IS '删除时间';

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

COMMENT ON TABLE plugin_marketplace_plugin_tag IS '插件市场插件标签关联表';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."plugin_record_id" IS '归属市场插件记录 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."tag_code" IS '稳定标签编码';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_plugin_tag."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_plugin_tag ON plugin_marketplace_plugin_tag ("plugin_record_id", "tag_code");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_tag_code ON plugin_marketplace_plugin_tag ("tag_code", "plugin_record_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_plugin_tag_plugin ON plugin_marketplace_plugin_tag ("plugin_id", "tag_code");

-- ============================================================
-- 私有插件与预留授权插件的可见性授权
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

COMMENT ON TABLE plugin_marketplace_visibility_grant IS '插件市场可见性授权表';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."plugin_record_id" IS '归属市场插件记录 ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."scope_type" IS '可见性范围类型：public/tenant/org/user/reserved_license';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."scope_id" IS '范围标识，public 范围时为空';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."permission" IS '授权覆盖的权限：view/download';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."status" IS '状态：0=停用，1=启用';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."expires_at" IS '授权过期时间';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_visibility_grant."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_visibility_grant ON plugin_marketplace_visibility_grant ("plugin_record_id", "scope_type", "scope_id", "permission");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_visibility_lookup ON plugin_marketplace_visibility_grant ("scope_type", "scope_id", "permission", "status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_visibility_plugin ON plugin_marketplace_visibility_grant ("plugin_id", "permission", "status");

-- ============================================================
-- 短期下载会话与授权快照
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

COMMENT ON TABLE plugin_marketplace_download_session IS '插件市场下载会话表';
COMMENT ON COLUMN plugin_marketplace_download_session."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."session_id" IS '不透明下载会话 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."release_id" IS '归属版本 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_id" IS '归属产物 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_download_session."requester_user_id" IS '请求用户 ID';
COMMENT ON COLUMN plugin_marketplace_download_session."status" IS '会话状态：active/expired/consumed/revoked';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_type" IS '会话绑定的产物类型';
COMMENT ON COLUMN plugin_marketplace_download_session."artifact_size_bytes" IS '产物大小（字节）';
COMMENT ON COLUMN plugin_marketplace_download_session."sha256" IS '返回给客户端的产物 SHA-256 校验和';
COMMENT ON COLUMN plugin_marketplace_download_session."authorization_snapshot" IS '创建会话时捕获的授权决策快照';
COMMENT ON COLUMN plugin_marketplace_download_session."expires_at" IS '会话过期时间';
COMMENT ON COLUMN plugin_marketplace_download_session."consumed_at" IS '首次成功下载时间';
COMMENT ON COLUMN plugin_marketplace_download_session."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_download_session."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_download_session."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_download_session_id ON plugin_marketplace_download_session ("session_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_expire ON plugin_marketplace_download_session ("status", "expires_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_requester ON plugin_marketplace_download_session ("requester_user_id", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_session_release ON plugin_marketplace_download_session ("release_id", "status");

-- ============================================================
-- 下载事件（异步聚合到读模型计数）
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

COMMENT ON TABLE plugin_marketplace_download_event IS '插件市场下载事件表';
COMMENT ON COLUMN plugin_marketplace_download_event."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."session_id" IS '不透明下载会话 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."release_id" IS '归属版本 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."artifact_id" IS '归属产物 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."release_version" IS '插件发布版本号';
COMMENT ON COLUMN plugin_marketplace_download_event."requester_user_id" IS '请求用户 ID';
COMMENT ON COLUMN plugin_marketplace_download_event."event_type" IS '下载事件类型：created/started/completed/failed';
COMMENT ON COLUMN plugin_marketplace_download_event."client_ip_hash" IS '用于统计的客户端 IP 哈希';
COMMENT ON COLUMN plugin_marketplace_download_event."user_agent_hash" IS '用于统计的 User-Agent 哈希';
COMMENT ON COLUMN plugin_marketplace_download_event."created_at" IS '创建时间';

CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_release ON plugin_marketplace_download_event ("release_id", "event_type", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_plugin ON plugin_marketplace_download_event ("plugin_id", "created_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_download_event_session ON plugin_marketplace_download_event ("session_id", "event_type");

-- ============================================================
-- 市场分页目录接口使用的列表投影
-- ============================================================
CREATE TABLE IF NOT EXISTS plugin_marketplace_plugin_read_model (
    "id"                 INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "plugin_record_id"   INT NOT NULL DEFAULT 0,
    "publisher_id"       INT NOT NULL DEFAULT 0,
    "publisher_name"     VARCHAR(128) NOT NULL DEFAULT '',
    "publisher_verified" BOOL NOT NULL DEFAULT FALSE,
    "plugin_id"          VARCHAR(64) NOT NULL DEFAULT '',
    "name"               VARCHAR(128) NOT NULL DEFAULT '',
    "summary"            VARCHAR(512) NOT NULL DEFAULT '',
    "plugin_type"        VARCHAR(32) NOT NULL DEFAULT 'source',
    "market_status"      VARCHAR(32) NOT NULL DEFAULT 'draft',
    "visibility"         VARCHAR(32) NOT NULL DEFAULT 'public',
    "latest_release_id"  INT NOT NULL DEFAULT 0,
    "latest_version"     VARCHAR(32) NOT NULL DEFAULT '',
    "min_host_version"   VARCHAR(32) NOT NULL DEFAULT '',
    "max_host_version"   VARCHAR(32) NOT NULL DEFAULT '',
    "primary_tag"        VARCHAR(64) NOT NULL DEFAULT '',
    "tag_codes"          JSONB NOT NULL DEFAULT '[]'::JSONB,
    "risk_counts"        JSONB NOT NULL DEFAULT '{}'::JSONB,
    "download_count"     BIGINT NOT NULL DEFAULT 0,
    "published_at"       TIMESTAMPTZ,
    "search_text"        TEXT,
    "created_at"         TIMESTAMPTZ,
    "updated_at"         TIMESTAMPTZ,
    "deleted_at"         TIMESTAMPTZ
);

COMMENT ON TABLE plugin_marketplace_plugin_read_model IS '插件市场列表读模型表';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."id" IS '主键 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_record_id" IS '归属市场插件记录 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_id" IS '归属发布者 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_name" IS '发布者展示名称快照';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."publisher_verified" IS '发布者认证状态快照';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_id" IS '稳定插件 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."name" IS '插件展示名称';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."summary" IS '市场短简介';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."plugin_type" IS '插件类型：source/dynamic';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."market_status" IS '市场状态';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."visibility" IS '可见性策略';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."latest_release_id" IS '最新已发布版本 ID';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."latest_version" IS '最新已发布版本号';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."min_host_version" IS '最低兼容 LinaPro 宿主版本';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."max_host_version" IS '最高兼容 LinaPro 宿主版本';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."primary_tag" IS '主分类标签编码';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."tag_codes" IS '展示用标签编码快照';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."risk_counts" IS '按级别聚合的风险计数快照';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."download_count" IS '聚合下载次数快照';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."published_at" IS '最近发布时间';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."search_text" IS '目录检索用纯文本投影';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."created_at" IS '创建时间';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."updated_at" IS '更新时间';
COMMENT ON COLUMN plugin_marketplace_plugin_read_model."deleted_at" IS '删除时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_marketplace_read_model_plugin ON plugin_marketplace_plugin_read_model ("plugin_id");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_status_updated ON plugin_marketplace_plugin_read_model ("market_status", "updated_at");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_type_status ON plugin_marketplace_plugin_read_model ("plugin_type", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_primary_tag ON plugin_marketplace_plugin_read_model ("primary_tag", "market_status");
CREATE INDEX IF NOT EXISTS idx_plugin_marketplace_read_model_visibility ON plugin_marketplace_plugin_read_model ("visibility", "market_status");

-- ============================================================
-- 字典初始化数据
-- ============================================================
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场插件类型', 'plugin_marketplace_plugin_type', 1, 1, '插件市场插件类型列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场状态', 'plugin_marketplace_status', 1, 1, '插件市场状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场审核状态', 'plugin_marketplace_review_status', 1, 1, '插件市场审核状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场产物类型', 'plugin_marketplace_artifact_type', 1, 1, '插件市场产物类型列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场风险类型', 'plugin_marketplace_risk_type', 1, 1, '插件市场风险类型列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场可见性', 'plugin_marketplace_visibility', 1, 1, '插件市场可见性列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场下载会话状态', 'plugin_marketplace_download_session_status', 1, 1, '插件市场下载会话状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '插件市场处理状态', 'plugin_marketplace_process_status', 1, 1, '插件市场异步处理状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_plugin_type', '源码插件', 'source', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_plugin_type', '动态插件', 'dynamic', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', '草稿', 'draft', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', '已发布', 'published', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', '已下架', 'delisted', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_status', '已废弃', 'deprecated', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', '草稿', 'draft', 1, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', '已提交', 'submitted', 2, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', '审核中', 'reviewing', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', '已通过', 'approved', 4, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_review_status', '已驳回', 'rejected', 5, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', '源码 ZIP', 'source_zip', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', '动态 ZIP', 'dynamic_zip', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_artifact_type', 'Plugin WASM', 'plugin_wasm', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '宿主服务', 'host_service', 1, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '动态路由', 'dynamic_route', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '菜单权限', 'menu_permission', 3, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '外部网络', 'external_network', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '数据表', 'data_table', 5, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '安装 SQL', 'install_sql', 6, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '卸载 SQL', 'uninstall_sql', 7, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', 'Mock SQL', 'mock_sql', 8, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '依赖', 'dependency', 9, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '多租户', 'multi_tenant', 10, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_risk_type', '文档', 'docs', 11, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', '公开', 'public', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', '私有', 'private', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_visibility', '预留授权', 'reserved', 3, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', '有效', 'active', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', '已过期', 'expired', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', '已消费', 'consumed', 3, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_download_session_status', '已撤销', 'revoked', 4, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_process_status', '待验证', 'pending_verify', 1, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_process_status', '待审核', 'pending_review', 2, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_process_status', '已完成', 'completed', 3, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'plugin_marketplace_process_status', '失败', 'failed', 4, 'error', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
