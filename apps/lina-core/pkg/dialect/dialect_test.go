// This file registers database drivers shared by dialect package tests.

package dialect

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"lina-core/internal/utility/testsupport"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPostgreSQLProjectSQLAssetsSmoke executes all PostgreSQL-source SQL
// assets against a real PostgreSQL database when explicitly enabled.
func TestPostgreSQLProjectSQLAssetsSmoke(t *testing.T) {
	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL SQL asset smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root failed: %v", err)
	}
	if !testsupport.OfficialPluginsWorkspaceReady(repoRoot) {
		t.Skip("official plugin workspace is not initialized")
	}

	var (
		installAssets   = collectProjectSQLAssets(t, sqlAssetGroupInstall)
		mockAssets      = collectProjectSQLAssets(t, sqlAssetGroupMock)
		uninstallAssets = collectProjectSQLAssets(t, sqlAssetGroupUninstall)
	)
	if len(installAssets) == 0 || len(mockAssets) == 0 || len(uninstallAssets) == 0 {
		t.Fatal("expected install, mock, and uninstall SQL assets")
	}

	dbLink := postgresSmokeDatabaseLink(t, baseLink)
	dbDialect, err := From(dbLink)
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, dbLink, true); err != nil {
		t.Fatalf("prepare PostgreSQL smoke database failed: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if cleanupErr := dropPostgreSQLSmokeDatabase(cleanupCtx, dbLink); cleanupErr != nil {
			t.Errorf("cleanup PostgreSQL smoke database failed: %v", cleanupErr)
		}
	})

	db, err := gdb.New(gdb.ConfigNode{Link: dbLink})
	if err != nil {
		t.Fatalf("open PostgreSQL smoke database failed: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(context.Background()); closeErr != nil {
			t.Errorf("close PostgreSQL smoke database failed: %v", closeErr)
		}
	})

	executePostgreSQLSQLAssets(t, ctx, db, installAssets)
	executePostgreSQLSQLAssets(t, ctx, db, mockAssets)
	executePostgreSQLSQLAssets(t, ctx, db, uninstallAssets)
}

// TestPostgreSQLProjectSQLUsesTimestamptzForInstants verifies host and plugin
// SQL assets create absolute instant columns as TIMESTAMPTZ.
func TestPostgreSQLProjectSQLUsesTimestamptzForInstants(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	removedMigrationPath := filepath.Join(root, "apps/lina-core/manifest/sql/013-timezone-aware-timestamps.sql")
	if _, err := os.Stat(removedMigrationPath); err == nil {
		t.Fatalf("core schema should not rely on standalone timestamp migration %s", removedMigrationPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat timestamp migration path: %v", err)
	}

	var files []string
	for _, group := range []sqlAssetGroup{sqlAssetGroupInstall, sqlAssetGroupMock} {
		for _, pattern := range sqlAssetPatterns(root, group) {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				t.Fatalf("glob project SQL assets failed: %v", err)
			}
			files = append(files, matches...)
		}
	}
	if len(files) == 0 {
		t.Fatal("expected project install or mock SQL assets")
	}

	bareTimestampType := regexp.MustCompile(`\bTIMESTAMP\b`)
	var (
		hostCombined   strings.Builder
		pluginCombined strings.Builder
	)
	for _, file := range files {
		content, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatalf("read project SQL asset %s: %v", file, readErr)
		}
		if match := bareTimestampType.FindStringIndex(string(content)); match != nil {
			t.Fatalf("project SQL asset %s still declares bare TIMESTAMP", file)
		}
		target := &hostCombined
		if strings.Contains(filepath.ToSlash(file), "/apps/lina-plugins/") {
			target = &pluginCombined
		}
		target.Write(content)
		target.WriteString("\n")
	}

	assertRequiredSQLPatterns(t, "host SQL assets", hostCombined.String(), []string{
		`"created_at"\s+TIMESTAMPTZ`,
		`"updated_at"\s+TIMESTAMPTZ`,
		`"start_at"\s+TIMESTAMPTZ`,
		`"read_at"\s+TIMESTAMPTZ`,
		`"expire_time"\s+TIMESTAMPTZ`,
		`"last_heartbeat_at"\s+TIMESTAMPTZ`,
		`"login_time"\s+TIMESTAMPTZ`,
	})
	if pluginCombined.Len() == 0 {
		return
	}
	pluginSQL := pluginCombined.String()
	pluginPatterns := make([]string, 0, 4)
	for _, column := range []string{"oper_time", "joined_at", "last_test_at", "expires_at"} {
		if strings.Contains(pluginSQL, `"`+column+`"`) {
			pluginPatterns = append(pluginPatterns, `"`+column+`"\s+TIMESTAMPTZ`)
		}
	}
	if len(pluginPatterns) == 0 {
		return
	}
	assertRequiredSQLPatterns(t, "plugin SQL assets", pluginSQL, pluginPatterns)
}

// executePostgreSQLSQLAssets executes assets in order on one PostgreSQL DB.
func executePostgreSQLSQLAssets(t *testing.T, ctx context.Context, db gdb.DB, assets []sqlTestAsset) {
	t.Helper()
	for _, asset := range assets {
		asset := asset
		t.Run(asset.sourceName, func(t *testing.T) {
			for index, statement := range SplitSQLStatements(asset.content) {
				if _, err := db.Exec(ctx, statement); err != nil {
					t.Fatalf("execute PostgreSQL statement %d failed: %v\nSQL:\n%s", index+1, err, statement)
				}
			}
		})
	}
}

// assertRequiredSQLPatterns verifies that one SQL asset group keeps representative
// instant columns on TIMESTAMPTZ.
func assertRequiredSQLPatterns(t *testing.T, group string, content string, requiredPatterns []string) {
	t.Helper()
	for _, required := range requiredPatterns {
		if !regexp.MustCompile(required).MatchString(content) {
			t.Fatalf("%s are missing %q", group, required)
		}
	}
}

// sqlAssetGroup identifies the lifecycle slice of project SQL assets.
type sqlAssetGroup string

const (
	sqlAssetGroupInstall   sqlAssetGroup = "install"
	sqlAssetGroupMock      sqlAssetGroup = "mock"
	sqlAssetGroupUninstall sqlAssetGroup = "uninstall"
)

// sqlAssetPatterns returns the project SQL glob patterns for one asset group.
func sqlAssetPatterns(root string, group sqlAssetGroup) []string {
	switch group {
	case sqlAssetGroupInstall:
		return []string{
			filepath.Join(root, "apps/lina-core/manifest/sql/*.sql"),
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/*.sql"),
		}
	case sqlAssetGroupMock:
		return []string{
			filepath.Join(root, "apps/lina-core/manifest/sql/mock-data/*.sql"),
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/mock-data/*.sql"),
		}
	case sqlAssetGroupUninstall:
		return []string{
			filepath.Join(root, "apps/lina-plugins/*/manifest/sql/uninstall/*.sql"),
		}
	default:
		return nil
	}
}

// sqlTestAsset stores one SQL asset fixture.
type sqlTestAsset struct {
	sourceName string
	content    string
}

// collectProjectSQLAssets finds host and plugin SQL assets relative to the repo root.
func collectProjectSQLAssets(t *testing.T, group sqlAssetGroup) []sqlTestAsset {
	t.Helper()
	root := repositoryRootForSQLAudit(t)
	var assets []sqlTestAsset
	for _, pattern := range sqlAssetPatterns(root, group) {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob SQL assets failed: %v", err)
		}
		for _, file := range files {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatalf("read SQL asset %s failed: %v", file, readErr)
			}
			rel, relErr := filepath.Rel(root, file)
			if relErr != nil {
				t.Fatalf("rel SQL asset %s failed: %v", file, relErr)
			}
			assets = append(assets, sqlTestAsset{
				sourceName: filepath.ToSlash(rel),
				content:    string(content),
			})
		}
	}
	return assets
}

// postgresSmokeDatabaseLink returns a unique database link for one smoke test.
func postgresSmokeDatabaseLink(t *testing.T, baseLink string) string {
	t.Helper()

	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		t.Fatalf("parse PostgreSQL smoke base link failed: %v", err)
	}
	if db.GetConfig() == nil {
		t.Fatal("PostgreSQL smoke base link configuration is empty")
	}
	config := db.GetConfig()
	if closeErr := db.Close(context.Background()); closeErr != nil {
		t.Fatalf("close PostgreSQL smoke base link parser failed: %v", closeErr)
	}

	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_sql_smoke_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		extra,
	)
}

// dropPostgreSQLSmokeDatabase removes the temporary database created by the
// PostgreSQL asset smoke test.
func dropPostgreSQLSmokeDatabase(ctx context.Context, targetLink string) (err error) {
	targetDB, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	targetConfig := targetDB.GetConfig()
	if targetConfig == nil {
		if closeErr := targetDB.Close(ctx); closeErr != nil {
			return closeErr
		}
		return nil
	}
	targetName := strings.TrimSpace(targetConfig.Name)
	if closeErr := targetDB.Close(ctx); closeErr != nil {
		return closeErr
	}
	if targetName == "" {
		return nil
	}

	systemLink := postgresSmokeSystemDatabaseLink(*targetConfig)
	systemDB, err := gdb.New(gdb.ConfigNode{Link: systemLink})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = systemDB.Exec(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		targetName,
	); err != nil {
		return err
	}
	quotedName := `"` + strings.ReplaceAll(targetName, `"`, `""`) + `"`
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// postgresSmokeSystemDatabaseLink returns a link to PostgreSQL's maintenance
// database using the same host, credentials, and extra parameters.
func postgresSmokeSystemDatabaseLink(config gdb.ConfigNode) string {
	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		extra,
	)
}

// TestFromResolvesPostgreSQL verifies pgsql: links dispatch to the PostgreSQL
// dialect and preserve PostgreSQL-source DDL.
func TestFromResolvesPostgreSQL(t *testing.T) {
	t.Parallel()

	dbDialect, err := From("pgsql:postgres:secret@tcp(127.0.0.1:5432)/linapro?sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if dbDialect.Name() != "postgres" {
		t.Fatalf("expected postgres dialect, got %s", dbDialect.Name())
	}
	if !dbDialect.SupportsCluster() {
		t.Fatal("expected PostgreSQL dialect to support cluster mode")
	}

	ddl := "CREATE TABLE demo (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"
	translated, err := dbDialect.TranslateDDL(context.Background(), "demo.sql", ddl)
	if err != nil {
		t.Fatalf("translate PostgreSQL DDL failed: %v", err)
	}
	if translated != ddl {
		t.Fatalf("expected PostgreSQL DDL to be unchanged, got %q", translated)
	}
}

// TestFromDriverTypeResolvesPostgreSQL verifies callers that only have a
// GoFrame driver type can still resolve the shared dialect boundary.
func TestFromDriverTypeResolvesPostgreSQL(t *testing.T) {
	t.Parallel()

	dbDialect, err := FromDriverType(" pgsql ")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect by driver type failed: %v", err)
	}
	if dbDialect.Name() != "postgres" {
		t.Fatalf("expected postgres dialect, got %s", dbDialect.Name())
	}
}

// TestFromRejectsMySQL verifies MySQL links fail with the explicit removal
// error instead of falling through to the generic unsupported-dialect path.
func TestFromRejectsMySQL(t *testing.T) {
	t.Parallel()

	_, err := From("mysql:root:secret@tcp(127.0.0.1:3306)/linapro")
	if err == nil {
		t.Fatal("expected MySQL dialect to be rejected")
	}
	if !bizerr.Is(err, CodeDialectMySQLUnsupported) {
		t.Fatalf("expected MySQL unsupported business error, got %v", err)
	}
}

// TestFromRejectsSQLite verifies SQLite links fail with the explicit removal
// error instead of falling through to the generic unsupported-dialect path.
func TestFromRejectsSQLite(t *testing.T) {
	t.Parallel()

	_, err := From("sqlite::@file(./temp/sqlite/linapro.db)")
	if err == nil {
		t.Fatal("expected SQLite dialect to be rejected")
	}
	if !bizerr.Is(err, CodeDialectSQLiteUnsupported) {
		t.Fatalf("expected SQLite unsupported business error, got %v", err)
	}
}

// TestFromRejectsPostgresAlias verifies only GoFrame's pgsql: prefix is
// accepted for PostgreSQL links.
func TestFromRejectsPostgresAlias(t *testing.T) {
	t.Parallel()

	if _, err := From("postgres:postgres://localhost/linapro"); err == nil {
		t.Fatal("expected postgres: alias to be unsupported")
	}
}

// TestOnConflictTargetsHaveDeclaredIdempotencyBasis verifies every SQL target
// using ON CONFLICT DO NOTHING has an explicit business-key or history policy.
func TestOnConflictTargetsHaveDeclaredIdempotencyBasis(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	sqlRoots := []string{
		filepath.Join(root, "apps", "lina-core", "manifest", "sql"),
		filepath.Join(root, "apps", "lina-plugins"),
	}
	targets := collectOnConflictTargets(t, sqlRoots...)

	allowedTargets := map[string]string{
		"plugin_linapro_content_notice":      "uk_plugin_linapro_content_notice_title(title)",
		"plugin_linapro_demo_dynamic_record": "primary key (id)",
		"plugin_linapro_demo_source_record":  "uk_plugin_linapro_demo_source_record_title(title)",
		"plugin_linapro_monitor_server":      "uk_plugin_linapro_monitor_server_node(node_name,node_ip)",
		"plugin_linapro_tenant_core_tenant":  "uk_plugin_linapro_tenant_core_tenant_code(code)",
		"plugin_linapro_org_core_dept":       "uk_plugin_linapro_org_core_dept_code(NULLIF(code,''))",
		"plugin_linapro_org_core_post":       "uk_plugin_linapro_org_core_post_code(code)",
		"plugin_linapro_org_core_user_dept":  "primary key (user_id,dept_id)",
		"plugin_linapro_org_core_user_post":  "primary key (user_id,post_id)",
		"sys_config":                         "uk_sys_config_tenant_key(tenant_id,key)",
		"sys_dict_data":                      "uk_sys_dict_data_tenant_type_value(tenant_id,dict_type,value)",
		"sys_dict_type":                      "uk_sys_dict_type_tenant_type(tenant_id,type)",
		"sys_job":                            "uk_sys_job_tenant_group_name(tenant_id,group_id,name)",
		"sys_job_group":                      "uk_sys_job_group_tenant_code(tenant_id,code)",
		"sys_menu":                           "uk_sys_menu_menu_key(menu_key)",
		"sys_notify_channel":                 "uk_sys_notify_channel_channel_key(channel_key)",
		"sys_online_session":                 "primary key (tenant_id,token_id)",
		"sys_role":                           "uk_sys_role_tenant_key(tenant_id,key)",
		"sys_role_menu":                      "primary key (role_id,menu_id,tenant_id)",
		"sys_user":                           "uk_sys_user_username(username)",
		"sys_user_role":                      "primary key (user_id,role_id,tenant_id)",
	}
	for _, historyTarget := range []string{
		"plugin_linapro_monitor_loginlog",
		"plugin_linapro_monitor_operlog",
		"sys_job_log",
		"sys_notify_delivery",
		"sys_notify_message",
	} {
		if containsString(targets, historyTarget) {
			t.Fatalf("history table %s must not use ON CONFLICT DO NOTHING", historyTarget)
		}
	}

	var missing []string
	for _, target := range targets {
		if allowedTargets[target] == "" {
			missing = append(missing, target)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("ON CONFLICT targets lack declared idempotency basis: %s", strings.Join(missing, ", "))
	}
}

// TestStaticHistoryMockInsertsHaveExistenceGuards verifies append-capable
// history tables keep production semantics while remaining mock-idempotent.
func TestStaticHistoryMockInsertsHaveExistenceGuards(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	sqlRoots := []string{
		filepath.Join(root, "apps", "lina-core", "manifest", "sql", "mock-data"),
		filepath.Join(root, "apps", "lina-plugins"),
	}
	historyTargets := map[string]struct{}{
		"plugin_linapro_monitor_loginlog": {},
		"plugin_linapro_monitor_operlog":  {},
		"sys_job_log":                     {},
		"sys_notify_delivery":             {},
		"sys_notify_message":              {},
	}
	statements := collectInsertStatementsForTargets(t, historyTargets, sqlRoots...)
	if len(statements) == 0 {
		t.Fatal("expected static history mock insert statements to be audited")
	}

	var missing []string
	notExistsPattern := regexp.MustCompile(`(?is)\bNOT\s+EXISTS\b`)
	for _, statement := range statements {
		if !notExistsPattern.MatchString(statement.sql) {
			missing = append(missing, statement.path+":"+statement.target)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("static history mock inserts lack WHERE NOT EXISTS guards: %s", strings.Join(missing, ", "))
	}
}

// TestSQLColumnIdentifiersAreQuoted verifies project SQL assets consistently
// wrap field identifiers with PostgreSQL double quotes.
func TestSQLColumnIdentifiersAreQuoted(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	sqlRoots := []string{
		filepath.Join(root, "apps", "lina-core", "manifest", "sql"),
		filepath.Join(root, "apps", "lina-plugins"),
	}
	columns := collectSQLColumnNames(t, sqlRoots...)
	if len(columns) == 0 {
		t.Fatal("expected SQL column definitions to be audited")
	}

	violations := collectUnquotedSQLColumnIdentifiers(t, columns, sqlRoots...)
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("SQL column identifiers must use PostgreSQL double quotes:\n%s", strings.Join(violations, "\n"))
	}
}

// TestSQLCreateTablesHaveBilingualPurposeComments verifies each table declares
// its business purpose directly above the CREATE TABLE statement.
func TestSQLCreateTablesHaveBilingualPurposeComments(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	sqlRoots := []string{
		filepath.Join(root, "apps", "lina-core", "manifest", "sql"),
		filepath.Join(root, "apps", "lina-plugins"),
	}

	violations := collectSQLCreateTablePurposeCommentViolations(t, sqlRoots...)
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("CREATE TABLE statements must have English and Chinese purpose comments:\n%s", strings.Join(violations, "\n"))
	}
}

// TestSeedDictDataKeepsValueAndLabel verifies seed dictionary rows still have
// a dict type, value, and label. Tag style uniqueness is not a host integrity rule.
func TestSeedDictDataKeepsValueAndLabel(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	sqlRoots := []string{
		filepath.Join(root, "apps", "lina-core", "manifest", "sql"),
		filepath.Join(root, "apps", "lina-plugins"),
	}
	entries := collectSeedDictDataTagStyles(t, sqlRoots...)
	if len(entries) == 0 {
		t.Fatal("expected seed dictionary data rows to be audited")
	}
	for _, entry := range entries {
		if entry.dictType == "" {
			t.Fatalf("seed dictionary row missing dict_type at %s", entry.location())
		}
	}
}

// TestPluginStateKeepsTechnicalPrimaryKeyAndBusinessUniqueIndex verifies
// the plugin state table keeps the project-wide technical id primary key while
// preserving the business uniqueness needed by InsertIgnore upserts.
func TestPluginStateKeepsTechnicalPrimaryKeyAndBusinessUniqueIndex(t *testing.T) {
	root := repositoryRootForSQLAudit(t)
	content, err := os.ReadFile(filepath.Join(root, "apps", "lina-core", "manifest", "sql", "009-plugin-host-call.sql"))
	if err != nil {
		t.Fatalf("read plugin host-call SQL failed: %v", err)
	}

	sql := string(content)
	idPrimaryKeyPattern := regexp.MustCompile(`(?is)"id"\s+INT\s+GENERATED\s+ALWAYS\s+AS\s+IDENTITY\s+PRIMARY\s+KEY`)
	if !idPrimaryKeyPattern.MatchString(sql) {
		t.Fatal("sys_plugin_state must keep id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY")
	}

	businessUniquePattern := regexp.MustCompile(`(?is)CREATE\s+UNIQUE\s+INDEX\s+IF\s+NOT\s+EXISTS\s+uk_sys_plugin_state_plugin_tenant_key\s+ON\s+sys_plugin_state\s*\(\s*"plugin_id"\s*,\s*"tenant_id"\s*,\s*"state_key"\s*\)`)
	if !businessUniquePattern.MatchString(sql) {
		t.Fatal("sys_plugin_state must keep a unique index on plugin_id, tenant_id, state_key")
	}
}

// dictDataTagStyleEntry records one seed sys_dict_data tag-style assignment.
type dictDataTagStyleEntry struct {
	path     string
	line     int
	dictType string
	value    string
	label    string
	tagStyle string
}

// location returns a compact source location for duplicate diagnostics.
func (e dictDataTagStyleEntry) location() string {
	return e.path + ":" + strconv.Itoa(e.line) + " value=" + strconv.Quote(e.value) + " label=" + strconv.Quote(e.label)
}

// collectOnConflictTargets returns the unique INSERT targets whose statements
// contain ON CONFLICT DO NOTHING.
func collectOnConflictTargets(t *testing.T, roots ...string) []string {
	t.Helper()

	insertTargetPattern := regexp.MustCompile(`(?is)\bINSERT\s+INTO\s+("?[\w]+"?)\b.*?\bON\s+CONFLICT\s+DO\s+NOTHING\b`)
	targets := map[string]struct{}{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, match := range insertTargetPattern.FindAllStringSubmatch(string(content), -1) {
				targets[strings.Trim(strings.ToLower(match[1]), `"`)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}

	result := make([]string, 0, len(targets))
	for target := range targets {
		result = append(result, target)
	}
	sort.Strings(result)
	return result
}

// collectSeedDictDataTagStyles returns sys_dict_data values inserted by seed
// SQL files together with the tag style assigned to each dictionary value.
func collectSeedDictDataTagStyles(t *testing.T, roots ...string) []dictDataTagStyleEntry {
	t.Helper()

	var result []dictDataTagStyleEntry
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") || strings.Contains(filepath.ToSlash(path), "/mock-data/") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, statement := range splitSQLAuditStatements(string(content)) {
				if !isDictDataInsertStatement(statement) {
					continue
				}
				parsed, ok := parseDictDataInsertStatement(statement)
				if !ok {
					return fmt.Errorf("parse sys_dict_data seed insert %s failed", filepath.ToSlash(path))
				}
				parsed.path = filepath.ToSlash(path)
				parsed.line = lineNumberForStatement(string(content), statement)
				result = append(result, parsed)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}
	return result
}

// isDictDataInsertStatement reports whether statement inserts seed dictionary
// data.
func isDictDataInsertStatement(statement string) bool {
	insertPattern := regexp.MustCompile(`(?is)^\s*INSERT\s+INTO\s+sys_dict_data\b`)
	return insertPattern.MatchString(statement)
}

// parseDictDataInsertStatement extracts dictionary type, value, label, and tag
// style from the project's literal sys_dict_data seed inserts.
func parseDictDataInsertStatement(statement string) (dictDataTagStyleEntry, bool) {
	insertPattern := regexp.MustCompile(`(?is)^\s*INSERT\s+INTO\s+sys_dict_data\s*\((.*?)\)\s*VALUES\s*\((.*?)\)\s*(?:ON\s+CONFLICT\s+DO\s+NOTHING)?\s*$`)
	match := insertPattern.FindStringSubmatch(statement)
	if len(match) < 3 {
		return dictDataTagStyleEntry{}, false
	}

	columns := parseSQLIdentifierList(match[1])
	values := parseSQLLiteralList(match[2])
	if len(columns) != len(values) {
		return dictDataTagStyleEntry{}, false
	}

	row := make(map[string]string, len(columns))
	for index, column := range columns {
		row[column] = values[index]
	}
	dictType, hasDictType := row["dict_type"]
	tagStyle, hasTagStyle := row["tag_style"]
	if !hasDictType || !hasTagStyle {
		return dictDataTagStyleEntry{}, false
	}
	return dictDataTagStyleEntry{
		dictType: dictType,
		value:    row["value"],
		label:    row["label"],
		tagStyle: tagStyle,
	}, true
}

// parseSQLIdentifierList parses a comma-separated SQL identifier list.
func parseSQLIdentifierList(content string) []string {
	parts := strings.Split(content, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		identifier := strings.TrimSpace(part)
		identifier = strings.Trim(identifier, `"`)
		result = append(result, strings.ToLower(identifier))
	}
	return result
}

// parseSQLLiteralList parses the literal VALUES list shape used by seed SQL.
func parseSQLLiteralList(content string) []string {
	var values []string
	for index := 0; index < len(content); {
		for index < len(content) && (content[index] == ' ' || content[index] == '\n' || content[index] == '\r' || content[index] == '\t' || content[index] == ',') {
			index++
		}
		if index >= len(content) {
			break
		}
		if content[index] == '\'' {
			value, next := parseSQLStringLiteral(content, index)
			values = append(values, value)
			index = next
			continue
		}
		start := index
		for index < len(content) && content[index] != ',' {
			index++
		}
		values = append(values, strings.TrimSpace(content[start:index]))
	}
	return values
}

// parseSQLStringLiteral parses a single-quoted SQL string literal.
func parseSQLStringLiteral(content string, start int) (string, int) {
	var builder strings.Builder
	for index := start + 1; index < len(content); index++ {
		if content[index] != '\'' {
			builder.WriteByte(content[index])
			continue
		}
		if index+1 < len(content) && content[index+1] == '\'' {
			builder.WriteByte('\'')
			index++
			continue
		}
		return builder.String(), index + 1
	}
	return builder.String(), len(content)
}

// lineNumberForStatement returns the one-based line number where statement
// starts within content.
func lineNumberForStatement(content string, statement string) int {
	index := strings.Index(content, statement)
	if index < 0 {
		return 1
	}
	return strings.Count(content[:index], "\n") + 1
}

// collectSQLColumnNames returns all column names declared by CREATE TABLE
// statements in project SQL assets.
func collectSQLColumnNames(t *testing.T, roots ...string) map[string]struct{} {
	t.Helper()

	columnDefinitionPattern := regexp.MustCompile(`(?i)^\s*"?([A-Za-z_][A-Za-z0-9_]*)"?\s+(?:INT|BIGINT|SMALLINT|VARCHAR|CHAR|TEXT|BYTEA|TIMESTAMP|DECIMAL|NUMERIC|REAL|DOUBLE\s+PRECISION)\b`)
	columns := map[string]struct{}{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, statement := range splitSQLAuditStatements(string(content)) {
				statementBody := trimLeadingSQLLineComments(statement)
				if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(statementBody)), "CREATE TABLE") {
					continue
				}
				for _, line := range strings.Split(statementBody, "\n") {
					match := columnDefinitionPattern.FindStringSubmatch(line)
					if len(match) < 2 {
						continue
					}
					columns[strings.ToLower(match[1])] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}
	return columns
}

// collectUnquotedSQLColumnIdentifiers returns source locations where known
// column names appear as unquoted SQL tokens outside literals and comments.
func collectUnquotedSQLColumnIdentifiers(t *testing.T, columns map[string]struct{}, roots ...string) []string {
	t.Helper()

	identifierPattern := regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	allowedKeywordPhrasePattern := regexp.MustCompile(`(?i)\b(?:PRIMARY|FOREIGN|UNIQUE)\s+KEY\b`)
	var violations []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			masked := maskSQLQuotedAndCommentedText(string(content))
			lines := strings.Split(masked, "\n")
			for index, line := range lines {
				line = allowedKeywordPhrasePattern.ReplaceAllString(line, "")
				for _, token := range identifierPattern.FindAllString(line, -1) {
					if _, ok := columns[strings.ToLower(token)]; !ok {
						continue
					}
					violations = append(violations, filepath.ToSlash(path)+":"+strconv.Itoa(index+1)+": "+token)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}
	return violations
}

// trimLeadingSQLLineComments removes adjacent line comments before a SQL
// statement so audits still classify statements with governance comments.
func trimLeadingSQLLineComments(statement string) string {
	lines := strings.Split(statement, "\n")
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[0])
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			lines = lines[1:]
			continue
		}
		break
	}
	return strings.Join(lines, "\n")
}

// collectSQLCreateTablePurposeCommentViolations returns CREATE TABLE locations
// that lack adjacent English and Chinese purpose comments.
func collectSQLCreateTablePurposeCommentViolations(t *testing.T, roots ...string) []string {
	t.Helper()

	createTablePattern := regexp.MustCompile(`(?i)^\s*CREATE\s+TABLE\b`)
	var violations []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			lines := strings.Split(string(content), "\n")
			for index, line := range lines {
				if !createTablePattern.MatchString(line) {
					continue
				}
				comments := collectPrecedingSQLCommentBlock(lines, index)
				if sqlCommentBlockHasPurposeComments(comments) {
					continue
				}
				violations = append(violations, filepath.ToSlash(path)+":"+strconv.Itoa(index+1))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}
	return violations
}

// collectPrecedingSQLCommentBlock returns the adjacent line-comment block above
// one SQL statement, allowing blank spacer lines inside the block.
func collectPrecedingSQLCommentBlock(lines []string, statementLine int) []string {
	var comments []string
	for index := statementLine - 1; index >= 0; index-- {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "--") {
			comments = append([]string{trimmed}, comments...)
			continue
		}
		break
	}
	return comments
}

// sqlCommentBlockHasPurposeComments reports whether a comment block contains
// both the English and Chinese purpose markers used by SQL asset governance.
func sqlCommentBlockHasPurposeComments(comments []string) bool {
	const (
		englishPurposePrefix = "-- Purpose:"
		chinesePurposePrefix = "-- \u7528\u9014\uff1a"
	)

	var (
		hasEnglishPurpose bool
		hasChinesePurpose bool
	)
	for _, comment := range comments {
		if strings.HasPrefix(comment, englishPurposePrefix) {
			hasEnglishPurpose = true
		}
		if strings.HasPrefix(comment, chinesePurposePrefix) {
			hasChinesePurpose = true
		}
	}
	return hasEnglishPurpose && hasChinesePurpose
}

// auditedInsertStatement records one SQL INSERT statement for audit checks.
type auditedInsertStatement struct {
	path   string
	target string
	sql    string
}

// collectInsertStatementsForTargets returns INSERT statements targeting any of
// the requested tables.
func collectInsertStatementsForTargets(t *testing.T, targets map[string]struct{}, roots ...string) []auditedInsertStatement {
	t.Helper()

	insertTargetPattern := regexp.MustCompile(`(?is)\bINSERT\s+INTO\s+("?[\w]+"?)\b`)
	var result []auditedInsertStatement
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, statement := range splitSQLAuditStatements(string(content)) {
				match := insertTargetPattern.FindStringSubmatch(statement)
				if len(match) < 2 {
					continue
				}
				target := strings.Trim(strings.ToLower(match[1]), `"`)
				if _, ok := targets[target]; !ok {
					continue
				}
				result = append(result, auditedInsertStatement{
					path:   filepath.ToSlash(path),
					target: target,
					sql:    statement,
				})
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan SQL root %s failed: %v", root, err)
		}
	}
	return result
}

// maskSQLQuotedAndCommentedText replaces string literals, quoted identifiers,
// and comments with spaces while preserving line numbers for diagnostics.
func maskSQLQuotedAndCommentedText(content string) string {
	var builder strings.Builder
	for index := 0; index < len(content); {
		switch {
		case content[index] == '\'':
			index = maskQuotedSQLText(&builder, content, index, '\'')
		case content[index] == '"':
			index = maskQuotedSQLText(&builder, content, index, '"')
		case content[index] == '-' && index+1 < len(content) && content[index+1] == '-':
			index = maskSQLLineComment(&builder, content, index)
		case content[index] == '/' && index+1 < len(content) && content[index+1] == '*':
			index = maskSQLBlockComment(&builder, content, index)
		default:
			builder.WriteByte(content[index])
			index++
		}
	}
	return builder.String()
}

// maskQuotedSQLText masks one SQL string literal or quoted identifier.
func maskQuotedSQLText(builder *strings.Builder, content string, start int, quote byte) int {
	builder.WriteByte(' ')
	index := start + 1
	for index < len(content) {
		current := content[index]
		if current == '\n' {
			builder.WriteByte('\n')
			index++
			continue
		}
		builder.WriteByte(' ')
		if current == quote {
			if index+1 < len(content) && content[index+1] == quote {
				builder.WriteByte(' ')
				index += 2
				continue
			}
			return index + 1
		}
		index++
	}
	return index
}

// maskSQLLineComment masks one line comment.
func maskSQLLineComment(builder *strings.Builder, content string, start int) int {
	index := start
	for index < len(content) && content[index] != '\n' {
		builder.WriteByte(' ')
		index++
	}
	return index
}

// maskSQLBlockComment masks one block comment.
func maskSQLBlockComment(builder *strings.Builder, content string, start int) int {
	index := start
	for index < len(content) {
		if content[index] == '\n' {
			builder.WriteByte('\n')
			index++
			continue
		}
		builder.WriteByte(' ')
		if content[index] == '*' && index+1 < len(content) && content[index+1] == '/' {
			builder.WriteByte(' ')
			return index + 2
		}
		index++
	}
	return index
}

// splitSQLAuditStatements splits the project SQL subset into statements.
func splitSQLAuditStatements(content string) []string {
	parts := strings.Split(content, ";")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

// repositoryRootForSQLAudit finds the monorepo root from this package's test
// working directory.
func repositoryRootForSQLAudit(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("read test working directory failed: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.work")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root containing go.work")
		}
		dir = parent
	}
}

// containsString reports whether values contains target.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
