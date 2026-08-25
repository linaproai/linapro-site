## 1. 构造期绑定后端

- [x] 1.1 `locker.New`接收`LockStore`；删除生产路径`ConfigureCoordination`
- [x] 1.2 会话存储构造时绑定热状态后端与 DB 投影；删除生产路径`session.ConfigureCoordination`
- [x] 1.3 `cachecoord.New`接收拓扑与 coordination；删除生产路径`Default`/`DefaultWithCoordination`
- [x] 1.4 `config`构造注入同一`cachecoord`实例，去掉内部`Default`调用

## 2. 启动装配与测试夹具

- [x] 2.1 `httpstartup`一次性创建并下发锁、会话、缓存协调实例
- [x] 2.2 更新所有生产`New()`调用点与单测，改为显式传入假后端或真实后端
- [x] 2.3 增加「构造后再启动」不再改变后端的回归测试

## 3. 死路径对齐

- [x] 3.1 注释与域说明对齐 Redis/进程内修订，不再宣称`sys_cache_revision`为现行实现
- [x] 3.2 确认服务层零调用后，按数据库规则评估删除或停止维护该表

## 4. 验证

- [x] 4.1 运行锁、会话、缓存协调、配置与`internal/cmd`启动绑定测试
- [x] 4.2 记录 DI 来源：owner、创建位置、传递路径、是否复用启动期实例
- [x] 4.3 运行`openspec validate startup-di-backends --strict`

### 任务记录

- **DI 来源检查**：
  - Owner：`httpstartup.newHTTPRuntime`（HTTP 组合根）。
  - 创建位置：读完 `cluster.enabled` 与方言覆盖后，创建 `coordination.Service` 与 `cluster.Service`，再 `cachecoord.New(clusterSvc, coordinationSvc)`、`config.New(cacheCoordSvc)`、`locker.New(coordination.LockStore 或 nil)`、`session.NewCoordinationStore(coordinationSvc, session.NewDBStore())`。
  - 传递路径：同一 `cacheCoordSvc` 传给 `i18n`、`role`、`sysinfo`、`plugin.New` 与 plugin runtime 调和器；同一 `lockerSvc`/`sessionStore` 传给 auth/plugin/hostlock。
  - 共享实例：生产路径只构造一份 cachecoord/locker/session/config 运行期图；`role`/`config` 不再内部 `cachecoord.Default`。测试传入独立 `cachecoord.New`/`coordination.NewMemory` 或 `New(nil)` SQL 锁。
- **i18n**：无用户可见文案或 apidoc 源文本变更。
- **缓存一致性**：单机走进程内修订（`New(topology, nil)`）；集群走注入的 coordination/Redis revision。`sys_cache_revision` 已从运行路径和 DAO 生成列表移除。
- **数据权限**：无会话列表接口重划。
- **开发工具跨平台**：仅改 `hack/config.yaml` 的 dao `tables` 列表，无新脚本；Windows/Linux/macOS 共用同一 yaml。
- **测试策略**：单测覆盖构造隔离、锁/会话/cachecoord/config/`httpstartup` 路由装配。未新增 E2E（启动行为对用户不可见）。`go test ./internal/cmd` 中 `TestProductionPanicsMatchAllowlist` 仍指向 marketplace `plugin.go` init，与本变更无关。本地 `sysconfig` 若干用例因库缺 `system_manageable` 列失败，属环境 schema，非本变更引入。
- **SQL**：源码 SQL 不再创建 `sys_cache_revision`（从 `012-distributed-cache-consistency.sql` 拿掉建表，不另留 `DROP` 迁移文件）。服务层确认零调用后删除生成 DAO/DO/Entity。

## Feedback

- [x] **FB-1**: 不为已部署库保留 `sys_cache_revision` 的 DROP 迁移；源码 SQL 从不创建该表
