## 1. 数据模型

- [x] 1.1 在 `manifest/sql/001-add-plugin-marketplace.sql` 最终态中创建 `plugin_marketplace_display_i18n`（不单独增量 SQL；不建文档正文表）。
- [x] 1.2 更新插件 `hack/config.yaml` 的 dao tables，并维护 DAO/DO/Entity。

## 2. 扫描写入

- [x] 2.1 实现从源码包 `manifest/i18n/<locale>/*.json` 与动态包 runtime i18n 提取 `plugin.<id>.name` / `plugin.<id>.description` 的解析逻辑。
- [x] 2.2 在源码/动态包上传事务中替换该 release 的 display_i18n 行；无 i18n 时写入 plugin.yaml 默认 locale 回退行。
- [x] 2.3 单元测试：多 locale 合并、缺 i18n 回退、locale 选择链（`marketplace_display_i18n_test.go`）。

## 3. 读路径投影

- [x] 3.1 列表（owned/catalog/managed）与详情按请求 locale 批量加载 display_i18n 并回退投影 `name`/`summary`。
- [x] 3.2 控制器从 BizCtx/Accept-Language 解析 locale 传入列表/详情查询（`resolveDocumentLocale`）。
- [x] 3.3 单元测试：zh-CN / en-US / 缺失回退。

## 4. 前端

- [x] 4.1 「我的插件」页 watch `preferences.app.locale` 后重新 query。
- [x] 4.2 管理列表页同样在语言切换后重新 query（目录页已重定向到我的插件）。

## 5. 文档表退役（磁盘读路径）

- [x] 5.1 文档读路径改为包 ZIP / Git docs-snapshot 磁盘读取并按需渲染。
- [x] 5.2 删除 `replaceReleaseDocuments` 与上传路径对 doc 表的写入；Git enrichment 改为写磁盘快照。
- [x] 5.3 删除 `plugin_marketplace_doc` 表定义与 DAO/DO/Entity、写入路径；SQL 合并为单一最终态 `001`。

## 6. 验证与边界

- [x] 6.1 确认文档/图片正文不入库；仅 `plugin_marketplace_display_i18n` 存名称/摘要。
- [x] 6.2 运行 `GOWORK=off go test ./backend/... -count=1` 通过；OpenSpec 变更已更新；影响：`i18n` 有（展示投影）；缓存无新增；数据权限沿用既有列表可见性；开发工具跨平台无影响。

## Feedback

- [x] **FB-1**: Git 源同步（如 official-plugins monorepo）仅写入 plugin.yaml 单语展示行，未读取远程 `manifest/i18n/<locale>` 的 `plugin.<id>.name` / `description`，导致「我的插件」列表切换语言后名称与摘要不变
- [x] **FB-2**: 审核队列未传递请求 locale，也未按待审版本批量投影 display i18n，导致中文「插件审核」页仍显示英文插件名；补齐后端批量投影、三页语言切换刷新与中英文本地化回归

### FB-2 影响与验证记录

- `i18n`有影响：审核队列按请求`locale`投影`pluginName`，「我的插件」「插件列表」「插件审核」切换语言后重建筛选项与列标题并重新查询；后端单元测试、前端单元测试、`Marketplace E2E`与真实后端中英文切换均已覆盖。
- 缓存一致性无影响：未新增或修改缓存、快照、失效与跨实例同步逻辑，展示元数据仍直接读取既有权威表。
- 数据权限无边界变化：审核队列继续使用既有`market:plugin:review`权限与既有分页可见范围，只对已返回当前页版本追加展示字段投影。
- 开发工具跨平台无影响：未修改`Makefile`、`linactl`、构建、测试或脚本入口。
- 运行期依赖与`DI`无影响：未新增构造函数参数、服务依赖、启动装配或独立服务实例。
- 性能边界：当前页`release ID`通过一次批量`display i18n`查询装配，循环内只做内存映射与`locale`回退，不产生`N+1`查询。
- 已通过`GOWORK=off go test ./backend/... -count=1`、前端单测`10/10`、`vue-tsc`、`Marketplace E2E 9/9`、`make lint plugins=1`、`make plugins.check`、`Prettier`、`git diff --check`与`openspec validate marketplace-display-i18n-metadata --strict`。
- 全局`make i18n.check`已在应用当前工作区差异并按`linapro/main`锁定提交`ebbd06be350e1953a02cca39df6559eee0ad09d0`检出完整`official-plugins`的临时工作区通过；此前失败是当前官网检出仅包含`Marketplace`、未包含`linapro-content-notice`自有资源所致，三个宿主引用键在该插件`en-US`与`zh-CN`资源中均已存在。
