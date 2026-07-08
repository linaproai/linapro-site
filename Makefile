# LinaPro Framework - Root Makefile
# ===========================

BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
TEMP_DIR      := temp
PID_DIR       := $(TEMP_DIR)/pids
BACKEND_PID   := $(PID_DIR)/lina-core.pid
FRONTEND_PID  := $(PID_DIR)/lina-vben.pid
BACKEND_PORT  := 9120
FRONTEND_PORT := 5666
BACKEND_LOG   := $(TEMP_DIR)/lina-core.log
FRONTEND_LOG  := $(TEMP_DIR)/lina-vben.log
EMBED_DIR     := $(BACKEND_DIR)/internal/packed/public
OUTPUT_DIR    := $(TEMP_DIR)/output
LINACTL       := cd hack/tools/linactl && go run .

SITE_DIR := apps/lina-site
SITE_NAME := LinaPro official site
SITE_DEFAULT_HOST ?= 127.0.0.1
SITE_DEFAULT_PORT ?= 3000
SITE_DEFAULT_LOCALE ?= zh-Hans

ifneq ($(filter site.version,$(MAKECMDGOALS)),)
SITE_VERSION_ARG := $(word 2,$(MAKECMDGOALS))
ifneq ($(SITE_VERSION_ARG),)
.PHONY: $(SITE_VERSION_ARG)
$(SITE_VERSION_ARG):
	@:
endif
endif

# Include split makefile targets.
# 引入拆分后的 Makefile 目标文件。
include hack/makefiles/help.mk
include hack/makefiles/env.mk
include hack/makefiles/dev.mk
include hack/makefiles/build.mk
include hack/makefiles/plugins.mk
include hack/makefiles/image.mk
include hack/makefiles/release.mk
include hack/makefiles/lint.mk
include hack/makefiles/test.mk
include hack/makefiles/i18n.mk
include hack/makefiles/database.mk
include hack/makefiles/agents.mk

## site.dev: 启动官网本地开发服务
.PHONY: site.dev
site.dev:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	HOST="$(strip $(or $(host),$(HOST),$(SITE_DEFAULT_HOST)))"; \
	PORT="$(strip $(or $(port),$(PORT),$(SITE_DEFAULT_PORT)))"; \
	LOCALE="$(strip $(or $(locale),$(LOCALE),$(SITE_DEFAULT_LOCALE)))"; \
	if [ "$$LOCALE" = "zh-Hans" ]; then \
		START_SCRIPT="start-zh"; \
		SITE_PATH="/zh/"; \
	elif [ "$$LOCALE" = "en" ]; then \
		START_SCRIPT="start-en"; \
		SITE_PATH="/"; \
	else \
		echo "Error: unsupported locale '$$LOCALE' (supported: zh-Hans, en)"; \
		exit 1; \
	fi; \
	[ -f "$$SITE_DIR/package.json" ] || { echo "Error: missing $$SITE_DIR/package.json"; exit 1; }; \
	PACKAGE_MANAGER=""; \
	INSTALL_CMD=""; \
	START_CMD=""; \
	if [ -f "$$SITE_DIR/yarn.lock" ] && command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		START_CMD="yarn --cwd $$SITE_DIR run $$START_SCRIPT --host $$HOST --port $$PORT"; \
	elif [ -f "$$SITE_DIR/pnpm-lock.yaml" ] && command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		START_CMD="pnpm --dir $$SITE_DIR run $$START_SCRIPT -- --host $$HOST --port $$PORT"; \
	elif command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		START_CMD="yarn --cwd $$SITE_DIR run $$START_SCRIPT --host $$HOST --port $$PORT"; \
	elif command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		START_CMD="pnpm --dir $$SITE_DIR run $$START_SCRIPT -- --host $$HOST --port $$PORT"; \
	elif command -v npm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="npm"; \
		INSTALL_CMD="npm --prefix $$SITE_DIR install"; \
		START_CMD="npm --prefix $$SITE_DIR run $$START_SCRIPT -- --host $$HOST --port $$PORT"; \
	else \
		echo "Error: no supported package manager found (tried yarn, pnpm, npm)"; \
		exit 1; \
	fi; \
	if [ ! -x "$$SITE_DIR/node_modules/.bin/docusaurus" ]; then \
		echo "Installing $$SITE_NAME dependencies with $$PACKAGE_MANAGER..."; \
		eval "$$INSTALL_CMD"; \
	fi; \
	echo "Starting $$SITE_NAME at http://$$HOST:$$PORT$$SITE_PATH (locale=$$LOCALE, package-manager=$$PACKAGE_MANAGER)"; \
	eval "$$START_CMD"

## site.preview: 构建所有语言并启动官网预览服务
.PHONY: site.preview
site.preview:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	HOST="$(strip $(or $(host),$(HOST),$(SITE_DEFAULT_HOST)))"; \
	PORT="$(strip $(or $(port),$(PORT),$(SITE_DEFAULT_PORT)))"; \
	[ -f "$$SITE_DIR/package.json" ] || { echo "Error: missing $$SITE_DIR/package.json"; exit 1; }; \
	PACKAGE_MANAGER=""; \
	INSTALL_CMD=""; \
	BUILD_CMD=""; \
	SERVE_CMD=""; \
	if [ -f "$$SITE_DIR/yarn.lock" ] && command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		BUILD_CMD="yarn --cwd $$SITE_DIR build"; \
		SERVE_CMD="yarn --cwd $$SITE_DIR run serve --host $$HOST --port $$PORT"; \
	elif [ -f "$$SITE_DIR/pnpm-lock.yaml" ] && command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		BUILD_CMD="pnpm --dir $$SITE_DIR run build"; \
		SERVE_CMD="pnpm --dir $$SITE_DIR run serve -- --host $$HOST --port $$PORT"; \
	elif command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		BUILD_CMD="yarn --cwd $$SITE_DIR build"; \
		SERVE_CMD="yarn --cwd $$SITE_DIR run serve --host $$HOST --port $$PORT"; \
	elif command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		BUILD_CMD="pnpm --dir $$SITE_DIR run build"; \
		SERVE_CMD="pnpm --dir $$SITE_DIR run serve -- --host $$HOST --port $$PORT"; \
	elif command -v npm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="npm"; \
		INSTALL_CMD="npm --prefix $$SITE_DIR install"; \
		BUILD_CMD="npm --prefix $$SITE_DIR run build"; \
		SERVE_CMD="npm --prefix $$SITE_DIR run serve -- --host $$HOST --port $$PORT"; \
	else \
		echo "Error: no supported package manager found (tried yarn, pnpm, npm)"; \
		exit 1; \
	fi; \
	if [ ! -x "$$SITE_DIR/node_modules/.bin/docusaurus" ]; then \
		echo "Installing $$SITE_NAME dependencies with $$PACKAGE_MANAGER..."; \
		eval "$$INSTALL_CMD"; \
	fi; \
	echo "Building $$SITE_NAME (all locales) with $$PACKAGE_MANAGER..."; \
	eval "$$BUILD_CMD"; \
	echo "Starting preview server at http://$$HOST:$$PORT"; \
	echo "  English: http://$$HOST:$$PORT/"; \
	echo "  中文:    http://$$HOST:$$PORT/zh/"; \
	eval "$$SERVE_CMD"

## site.check: 检查官网中文文档在所有 i18n locale 中均有对应翻译文件
.PHONY: site.check
site.check:
	@bash .github/workflows/consistency-check.sh

## check: 兼容旧官网仓库的 i18n 完整性检查入口
.PHONY: check
check: site.check

## site.build: 编译生成官网静态文件
.PHONY: site.build
site.build:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	[ -f "$$SITE_DIR/package.json" ] || { echo "Error: missing $$SITE_DIR/package.json"; exit 1; }; \
	PACKAGE_MANAGER=""; \
	INSTALL_CMD=""; \
	BUILD_CMD=""; \
	if [ -f "$$SITE_DIR/yarn.lock" ] && command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		BUILD_CMD="yarn --cwd $$SITE_DIR build"; \
	elif [ -f "$$SITE_DIR/pnpm-lock.yaml" ] && command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		BUILD_CMD="pnpm --dir $$SITE_DIR run build"; \
	elif command -v yarn >/dev/null 2>&1; then \
		PACKAGE_MANAGER="yarn"; \
		INSTALL_CMD="yarn --cwd $$SITE_DIR install"; \
		BUILD_CMD="yarn --cwd $$SITE_DIR build"; \
	elif command -v pnpm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="pnpm"; \
		INSTALL_CMD="pnpm --dir $$SITE_DIR install"; \
		BUILD_CMD="pnpm --dir $$SITE_DIR run build"; \
	elif command -v npm >/dev/null 2>&1; then \
		PACKAGE_MANAGER="npm"; \
		INSTALL_CMD="npm --prefix $$SITE_DIR install"; \
		BUILD_CMD="npm --prefix $$SITE_DIR run build"; \
	else \
		echo "Error: no supported package manager found (tried yarn, pnpm, npm)"; \
		exit 1; \
	fi; \
	if [ ! -x "$$SITE_DIR/node_modules/.bin/docusaurus" ]; then \
		echo "Installing $$SITE_NAME dependencies with $$PACKAGE_MANAGER..."; \
		eval "$$INSTALL_CMD"; \
	fi; \
	echo "Building $$SITE_NAME with $$PACKAGE_MANAGER..."; \
	eval "$$BUILD_CMD"; \
	echo "Build complete: $$SITE_DIR/build/"

## site.version: 归档当前官网文档为指定版本，用法：make site.version 0.1.x
.PHONY: site.version
site.version:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	DOC_VERSION="$(strip $(or $(version),$(VERSION),$(SITE_VERSION_ARG)))"; \
	if [ -z "$$DOC_VERSION" ]; then \
		echo "Usage: make site.version <version>"; \
		exit 1; \
	fi; \
	[ -f "$$SITE_DIR/package.json" ] || { echo "Error: missing $$SITE_DIR/package.json"; exit 1; }; \
	if ! command -v yarn >/dev/null 2>&1; then \
		echo "Error: yarn is required"; \
		exit 1; \
	fi; \
	if [ ! -x "$$SITE_DIR/node_modules/.bin/docusaurus" ]; then \
		echo "Installing $$SITE_NAME dependencies with yarn..."; \
		yarn --cwd "$$SITE_DIR" install; \
	fi; \
	echo "Archiving $$SITE_NAME docs as version $$DOC_VERSION..."; \
	yarn --cwd "$$SITE_DIR" run docusaurus docs:version "$$DOC_VERSION"; \
	echo "Syncing versioned i18n docs for $$DOC_VERSION..."; \
	node "$$SITE_DIR/scripts/sync-versioned-i18n-docs.js" "$$DOC_VERSION"; \
	MAJOR=$$(echo "$$DOC_VERSION" | cut -d. -f1); \
	MINOR=$$(echo "$$DOC_VERSION" | cut -d. -f2); \
	NEXT_MINOR=$$((MINOR + 1)); \
	NEXT_VERSION="$${MAJOR}.$${NEXT_MINOR}.x"; \
	NEXT_LABEL="$${NEXT_VERSION}(Latest)"; \
	echo "Updating version label to $$NEXT_LABEL..."; \
	sed -i.bak "s|const LATEST_VERSION_LABEL = '[^']*';|const LATEST_VERSION_LABEL = '$$NEXT_LABEL';|" "$$SITE_DIR/docusaurus.config.ts" && rm -f "$$SITE_DIR/docusaurus.config.ts.bak"; \
	I18N_JSON="$$SITE_DIR/i18n/zh-Hans/docusaurus-plugin-content-docs/current.json"; \
	if [ -f "$$I18N_JSON" ]; then \
		sed -i.bak "s|\"message\": \"[^\"]*\",|\"message\": \"$$NEXT_LABEL\",|" "$$I18N_JSON" && rm -f "$$I18N_JSON.bak"; \
	fi; \
	echo "Version $$DOC_VERSION archived. Next development version: $$NEXT_LABEL"

## site.webp: 将官网内容图片转换为 WebP
WEBP_LOSSLESS ?= 1
WEBP_QUALITY ?= 100
WEBP_INCLUDE_STATIC ?= 0

.PHONY: site.webp
site.webp:
	cd $(SITE_DIR) && WEBP_LOSSLESS=$(WEBP_LOSSLESS) WEBP_QUALITY=$(WEBP_QUALITY) WEBP_INCLUDE_STATIC=$(WEBP_INCLUDE_STATIC) node scripts/docs-images-to-webp.js $(IMAGE_FLAGS)
