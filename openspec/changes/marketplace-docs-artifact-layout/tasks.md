## 1. ArtifactStore 删除能力

- [x] 1.1 `ArtifactStore` 接口增加幂等 `Delete(ctx, key)`
- [x] 1.2 `LocalArtifactStore` 实现 Delete，并补充单元测试
- [x] 1.3 测试用 mock（memory / uploadOwnership）补齐 Delete

## 2. 文档快照布局与同步

- [x] 2.1 调整 key 生成：`<plugin>/<version>/docs/<locale>/<docPath>` 与 `meta/docs-manifest.json`
- [x] 2.2 `replaceReleaseGitDocumentSnapshot`：hash 比对后按需 Put；manifest 记录 contentHash
- [x] 2.3 同步后按旧 manifest / 本地 docs 树清理孤儿文件
- [x] 2.4 读取路径 `loadDocumentIndexItemsFromGitSnapshot` 适配新 manifest 位置

## 3. 验证与文档

- [x] 3.1 更新 Git 文档索引单测：路径布局、hash 覆盖、未变跳过、孤儿删除
- [x] 3.2 更新 marketplace README / README.zh-CN 存储路径说明
- [x] 3.3 运行 marketplace 相关 `go test` 与 `openspec validate --strict`

### 影响分析

- i18n：无影响（无用户可见文案变更）
- 缓存一致性：有影响；Git 同步仍替换版本文档磁盘快照，详情 GET 只读快照不回源；同路径以 content hash 决定是否覆盖
- 数据权限：无影响
- DI：无新增运行期依赖
- 测试：单元测试覆盖布局、hash 更新、跳过、孤儿删除；无用户可观察 UI 变化，不新增 E2E
