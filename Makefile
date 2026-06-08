


SITE_DIR := apps/lina-site
SITE_NAME := LinaPro official site
SITE_DEFAULT_HOST ?= 127.0.0.1
SITE_DEFAULT_PORT ?= 3000
SITE_DEFAULT_LOCALE ?= zh-Hans

ifneq ($(filter version,$(MAKECMDGOALS)),)
VERSION_ARG := $(word 2,$(MAKECMDGOALS))
ifneq ($(VERSION_ARG),)
.PHONY: $(VERSION_ARG)
$(VERSION_ARG):
	@:
endif
endif

## dev: 启动官网本地开发服务
.PHONY: dev
dev:
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

## preview: 构建所有语言并启动预览服务（用于测试多语言切换）
.PHONY: preview
preview:
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

## check: 检查中文文档在所有 i18n locale 中均有对应翻译文件
.PHONY: check
check:
	@bash .github/workflows/consistency-check.sh

## build: 编译生成官网静态文件（输出到 apps/lina-site/build/）
.PHONY: build
build:
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

## version: 归档当前文档为指定版本，用法：make version 0.1.x
.PHONY: version
version:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	VERSION="$(VERSION_ARG)"; \
	if [ -z "$$VERSION" ]; then \
		echo "Usage: make version <version>"; \
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
	echo "Archiving $$SITE_NAME docs as version $$VERSION..."; \
	yarn --cwd "$$SITE_DIR" run docusaurus docs:version "$$VERSION"; \
	echo "Syncing versioned i18n docs for $$VERSION..."; \
	node "$$SITE_DIR/scripts/sync-versioned-i18n-docs.js" "$$VERSION"; \
	MAJOR=$$(echo "$$VERSION" | cut -d. -f1); \
	MINOR=$$(echo "$$VERSION" | cut -d. -f2); \
	NEXT_MINOR=$$((MINOR + 1)); \
	NEXT_VERSION="$${MAJOR}.$${NEXT_MINOR}.x"; \
	NEXT_LABEL="$${NEXT_VERSION}(Latest)"; \
	echo "Updating version label to $$NEXT_LABEL..."; \
	sed -i.bak "s|const LATEST_VERSION_LABEL = '[^']*';|const LATEST_VERSION_LABEL = '$$NEXT_LABEL';|" "$$SITE_DIR/docusaurus.config.ts" && rm -f "$$SITE_DIR/docusaurus.config.ts.bak"; \
	I18N_JSON="$$SITE_DIR/i18n/zh-Hans/docusaurus-plugin-content-docs/current.json"; \
	if [ -f "$$I18N_JSON" ]; then \
		sed -i.bak "s|\"message\": \"[^\"]*\",|\"message\": \"$$NEXT_LABEL\",|" "$$I18N_JSON" && rm -f "$$I18N_JSON.bak"; \
	fi; \
	echo "Version $$VERSION archived. Next development version: $$NEXT_LABEL"

## image: 将 docs/blog/i18n 中被引用的本地内容图片转换为 WebP，并更新引用、删除安全的原图 [IMAGE_FLAGS=--dry-run|--include-static] [WEBP_INCLUDE_STATIC=1|0] [WEBP_LOSSLESS=1|0] [WEBP_QUALITY=1-100]
# 默认使用无损 WebP，避免降低图片质量；只有 WEBP_LOSSLESS=0 时 WEBP_QUALITY 才用于有损压缩。
WEBP_LOSSLESS ?= 1
WEBP_QUALITY ?= 100
WEBP_INCLUDE_STATIC ?= 0

.PHONY: webp
webp:
	cd $(SITE_DIR) && WEBP_LOSSLESS=$(WEBP_LOSSLESS) WEBP_QUALITY=$(WEBP_QUALITY) WEBP_INCLUDE_STATIC=$(WEBP_INCLUDE_STATIC) node scripts/docs-images-to-webp.js $(IMAGE_FLAGS)
