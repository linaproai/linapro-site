


SITE_DIR := apps/lina-site
SITE_NAME := LinaPro official site
SITE_DEFAULT_HOST ?= 127.0.0.1
SITE_DEFAULT_PORT ?= 3000
SITE_DEFAULT_LOCALE ?= zh-Hans

## dev: 启动官网本地开发服务
.PHONY: dev
dev:
	@set -e; \
	SITE_DIR="$(SITE_DIR)"; \
	SITE_NAME="$(SITE_NAME)"; \
	HOST="$(strip $(or $(host),$(HOST),$(SITE_DEFAULT_HOST)))"; \
	PORT="$(strip $(or $(port),$(PORT),$(SITE_DEFAULT_PORT)))"; \
	LOCALE="$(strip $(or $(locale),$(LOCALE),$(SITE_DEFAULT_LOCALE)))"; \
	START_SCRIPT="start"; \
	if [ "$$LOCALE" = "en" ]; then \
		START_SCRIPT="start-en"; \
	elif [ "$$LOCALE" != "zh-Hans" ]; then \
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
	echo "Starting $$SITE_NAME at http://$$HOST:$$PORT (locale=$$LOCALE, package-manager=$$PACKAGE_MANAGER)"; \
	eval "$$START_CMD"

## image: 将 docs/blog/i18n 中被引用的本地内容图片转换为 WebP，并更新引用、删除安全的原图 [IMAGE_FLAGS=--dry-run|--include-static] [WEBP_INCLUDE_STATIC=1|0] [WEBP_LOSSLESS=1|0] [WEBP_QUALITY=1-100]
# 默认使用无损 WebP，避免降低图片质量；只有 WEBP_LOSSLESS=0 时 WEBP_QUALITY 才用于有损压缩。
WEBP_LOSSLESS ?= 1
WEBP_QUALITY ?= 100
WEBP_INCLUDE_STATIC ?= 0

.PHONY: image
image:
	cd $(SITE_DIR) && WEBP_LOSSLESS=$(WEBP_LOSSLESS) WEBP_QUALITY=$(WEBP_QUALITY) WEBP_INCLUDE_STATIC=$(WEBP_INCLUDE_STATIC) node scripts/docs-images-to-webp.js $(IMAGE_FLAGS)
