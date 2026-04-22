
.PHONY: dev

# 引用复杂指令子文件
include hack/makefiles/up.mk

## dev: 启动官网本地开发服务
dev:
	@set -e; \
	command -v pnpm >/dev/null 2>&1 || { echo "Error: 'pnpm' command not found"; exit 1; }; \
	if [ ! -x "apps/lina-site/node_modules/.bin/docusaurus" ]; then \
		echo "Installing workspace dependencies..."; \
		pnpm install; \
	fi; \
	echo "Starting LinaPro official site at http://localhost:3000"; \
	pnpm --filter @linapro/lina-site start


