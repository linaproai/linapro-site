## 1. 插件配置与manifest资源文档

- [x] 1.1 新增`9000-plugin-config-and-metadata.md`中文文档，补齐`front matter`、`description`和`keywords`
- [x] 1.2 说明插件配置文件职责、开发期路径、生产配置根路径、动态`artifact`默认配置和读取优先级
- [x] 1.3 说明`Config()`、`HostConfig()`、`Manifest()`职责边界，并提供插件个性化配置示例
- [x] 1.4 说明插件`manifest`资源的读取路径、专用管线边界和个性化使用方式

## 2. 现有插件文档入口调整

- [x] 2.1 更新源码插件文档中的配置与清单资源小节，保留摘要并链接到新增公共文档
- [x] 2.2 更新动态插件文档中的资源打包或目录结构说明，链接到新增公共文档
- [x] 2.3 检查插件系统相关中文文档是否需要补充入口，避免同一规则分散重复

## 3. AI项目规范管理文档

- [x] 3.1 扩写开发指令文档的`AI工具集成`章节，说明`make agents`、`skills`、`prompts`和`md`三类资源桥接策略
- [x] 3.2 扩写`AI原生设计`文档，说明`AGENTS.md`、`.agents/rules/*.md`、插件本地`AGENTS.md`和渐进式规范披露策略
- [x] 3.3 确认`apps/lina-site/i18n/`目录未因本次变更被修改

## 4. 验证

- [x] 4.1 按 Markdown 格式规范审查新增和修改的中文文档
- [x] 4.2 运行`openspec validate document-plugin-config-metadata-agents --strict`
- [x] 4.3 运行站点构建或等价文档检查，确认新增文档可被 Docusaurus 发现

## Feedback

- [x] **FB-1**: `9000-plugin-config-and-metadata.md`不应强调`metadata.yaml`具有特殊意义，应改为普通声明型`YAML`资源示例
- [x] **FB-2**: 根据最新`Manifest()`实现更新文档，说明其读取`manifest/`原始资源且不替代配置、`SQL`和`i18n`专用管线
