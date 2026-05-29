---
slug: '/docs/plugin-capability-services'
title: '插件可用基础能力概览'
hide_title: true
description: 'Services 接口的设计原则、服务分类、获取方式和按场景选服务指南，帮助插件开发者理解各能力服务的定位、边界和协作关系，建立正确的使用心智模型。'
keywords:
  - 插件能力
  - capability.Services
  - pluginhost.Services
  - 基础能力服务
  - 插件开发
  - 服务架构
  - 认证服务
  - 缓存服务
  - 配置服务
  - 国际化服务
  - 租户能力
  - 组织能力
  - 插件生命周期
  - 通知服务
  - 会话服务
  - LinaPro
---

## 基本介绍

`LinaPro`主框架通过`capability.Services`接口向插件暴露一组稳定的基础能力服务。这些服务覆盖了插件开发中最常见的横切关注点：认证与上下文、配置与资源、数据与存储、插件治理、通知、以及组织和租户等框架级能力。

源码插件通过`pluginhost.Services`获取完整的服务目录，它在`capability.Services`基础上扩展了`TenantFilter()`，为携带数据库查询构建器的能力提供了单独的入口。

这套服务架构遵循几个核心设计原则：

- **显式契约，稳定边界。** 每个服务都有明确的合约定义（`contract`包），插件只依赖稳定契约，不依赖宿主内部实现。
- **插件作用域隔离。** 配置、缓存、清单资源等服务自动绑定到当前插件ID，插件间不会互相干扰。
- **能力可选，安全降级。** 组织、租户等框架级能力在提供方不可用时自动降级，插件通过`Available()`检查可用性。
- **只读消费，最小暴露。** 普通插件获取的是只读消费接口，写操作和数据库查询构建器不通过`capability.Services`暴露。



## 能力速查

| 分类 | 服务 | 合约类型 | 简介 |
|------|------|----------|-----------|
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `APIDoc()` | `contract.APIDocService` | API文档本地化，解析路由操作键和翻译文本 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `Auth()` | `contract.AuthService` | 租户Token签发、切换和模拟令牌管理 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `BizCtx()` | `contract.BizCtxService` | 读取当前请求的用户、租户、模拟状态等业务上下文快照 |
| <span style={{whiteSpace: 'nowrap'}}>认证与上下文</span> | `I18n()` | `contract.I18nService` | 运行时翻译、获取请求Locale、搜索翻译键 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `Config()` | `contract.ConfigService` | 读取当前插件自己的静态配置 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `HostConfig()` | `contract.HostConfigService` | 读取宿主公开的配置白名单键 |
| <span style={{whiteSpace: 'nowrap'}}>配置与资源</span> | `Manifest()` | `contract.ManifestService` | 读取当前插件`manifest/`下的原始资源文件 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Cache()` | `contract.CacheService` | 插件作用域的运行时缓存 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Session()` | `contract.SessionService` | 在线会话管理：分页查询和踢出会话 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `Route()` | `contract.RouteService` | 获取当前动态路由的元数据 |
| <span style={{whiteSpace: 'nowrap'}}>插件治理</span> | `PluginLifecycle()` | `contract.PluginLifecycleService` | 插件生命周期编排：租户级禁用/删除的前置检查和通知 |
| <span style={{whiteSpace: 'nowrap'}}>插件治理</span> | `PluginState()` | `contract.PluginStateService` | 查询插件启用状态 |
| <span style={{whiteSpace: 'nowrap'}}>通知</span> | `Notify()` | `contract.NotifyService` | 发布通知到宿主收件箱 |
| <span style={{whiteSpace: 'nowrap'}}>能力提供方</span> | `Org()` | `orgcap.Service` | 组织能力消费：用户部门、岗位等只读投影 |
| <span style={{whiteSpace: 'nowrap'}}>能力提供方</span> | `Tenant()` | `tenantcap.Service` | 租户能力消费：当前租户、可见性校验、租户切换 |
| <span style={{whiteSpace: 'nowrap'}}>数据与存储</span> | `TenantFilter()` | `contract.TenantFilterService` | 为插件自有表注入`tenant_id`过滤条件 |

## 获取方式

### 源码插件

源码插件在路由注册、钩子回调和定时任务注册时，通过`registrar.Services()`获取完整的服务目录：

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.Services()

    // 通过 services 访问各能力服务
    config := services.Config()
    tenantFilter := services.TenantFilter()
    i18n := services.I18n()
    // ...
    return nil
}
```

`registrar.Services()`返回的`pluginhost.Services`嵌入了`capability.Services`的全部16个服务，并额外提供`TenantFilter()`。在钩子回调和定时任务注册中，同样通过`payload.Services()`或`registrar.Services()`获取。

### 动态插件

动态插件通过`pluginbridge`的`guest`侧`SDK`访问主框架能力。所有服务调用都经过`hostServices`桥接，必须先在`plugin.yaml`中声明授权：

```yaml
hostServices:
  - service: config
    methods: [get]
  - service: hostConfig
    methods: [get]
    resources:
      keys:
        - workspace.basePath
        - i18n.default
  - service: manifest
    methods: [get]
    resources:
      paths:
        - profile.yaml
  - service: cache
    methods: [get, set, delete, incr, expire]
  - service: i18n
    methods: [translate, getLocale]
  - service: notify
    methods: [send]
```

`guest`侧`SDK`提供的便捷函数（`Config.String`、`Config.Duration`、`I18n.Translate`、`Manifest.Scan`等）在运行时自动转化为对应的`hostServices`调用，不需要在`plugin.yaml`中把这些便捷函数名声明为`methods`。例如声明`methods: [get]`即可使用`Config.String`、`Config.Bool`、`Config.Int`、`Config.Duration`、`Config.Scan`等所有读取函数。

在`WASM`入口函数或控制器中，通过`pluginbridge`的`guest`侧能力获取各服务：

```go
// 通过 guest 侧 SDK 读取配置
endpoint, err := guestsdk.Config.String(ctx, "sync.endpoint", "")

// 通过 guest 侧 SDK 读取宿主配置
workspaceBase, err := guestsdk.HostConfig.String(ctx, "workspace.basePath", "/admin")

// 通过 guest 侧 SDK 读取 manifest 资源
var profile struct {
    Category string `yaml:"category"`
}
err := guestsdk.Manifest.Scan(ctx, "profile.yaml", "", &profile)

// 通过 guest 侧 SDK 读取翻译
message := guestsdk.I18n.Translate(ctx, "plugin.record.created", "Record created")

// 通过 hostServices 操作缓存
err := guestsdk.Cache.Set(ctx, "stats", "visit-count", "100", 0)

// 从 BridgeRequestEnvelopeV1 请求包中提取上下文
userID := requestEnvelope.UserID
tenantID := requestEnvelope.TenantID
```

动态插件的`BizCtx`上下文信息从`BridgeRequestEnvelopeV1`请求包中提取，而不是通过`services.BizCtx()`获取。请求包中的`UserID`、`TenantID`等字段由主框架在进入`WASM`沙箱前注入，语义与源码插件的`BizCtxService.Current()`一致。

## 代码示例

以下示例展示源码插件和动态插件使用各基础能力的常见模式。

### 读取配置

#### 源码插件

源码插件直接调用`services.Config()`和`services.HostConfig()`：

```go
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
    services := registrar.Services()

    syncEndpoint, err := services.Config().String(ctx, "sync.endpoint", "")
    if err != nil {
        return err
    }

    timeout, err := services.Config().Duration(ctx, "sync.timeout", 30*time.Second)
    if err != nil {
        return err
    }

    workspaceBase, err := services.HostConfig().String(ctx, "workspace.basePath", "/admin")
    if err != nil {
        return err
    }

    _ = syncEndpoint
    _ = timeout
    _ = workspaceBase
    return nil
}
```

#### 动态插件

动态插件通过`guest`侧`SDK`读取，需要在`plugin.yaml`中声明`config`和`hostConfig`授权：

```go
syncEndpoint, err := guestsdk.Config.String(ctx, "sync.endpoint", "")
if err != nil {
    return err
}

timeout, err := guestsdk.Config.Duration(ctx, "sync.timeout", 30*time.Second)
if err != nil {
    return err
}

workspaceBase, err := guestsdk.HostConfig.String(ctx, "workspace.basePath", "/admin")
if err != nil {
    return err
}
```

### 操作缓存

#### 源码插件

源码插件通过`services.Cache()`操作缓存：

```go
func (s *recordService) GetVisitCount(ctx context.Context, recordID int) (int, error) {
    services := s.services
    val, err := services.Cache().Get(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    if err != nil {
        return 0, err
    }
    count, _ := strconv.Atoi(val)
    return count, nil
}

func (s *recordService) IncrVisitCount(ctx context.Context, recordID int) error {
    services := s.services
    _, err := services.Cache().Incr(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    return err
}

func (s *recordService) CacheSummary(ctx context.Context, recordID int, summary string) error {
    services := s.services
    return services.Cache().Set(ctx, "summary", fmt.Sprintf("record:%d", recordID), summary, 10*time.Minute)
}
```

#### 动态插件

动态插件通过`hostServices`的`cache`授权操作缓存，需要在`plugin.yaml`中声明`service: cache`：

```go
func GetVisitCount(ctx context.Context, recordID int) (int, error) {
    val, err := guestsdk.Cache.Get(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    if err != nil {
        return 0, err
    }
    count, _ := strconv.Atoi(val)
    return count, nil
}

func IncrVisitCount(ctx context.Context, recordID int) error {
    _, err := guestsdk.Cache.Incr(ctx, "visit", fmt.Sprintf("record:%d", recordID))
    return err
}
```

### 国际化翻译

#### 源码插件

源码插件通过`services.I18n()`翻译文本：

```go
func (s *recordService) CreateRecord(ctx context.Context, req *CreateRecordReq) error {
    services := s.services

    // 插件语言包 manifest/i18n/zh-CN/plugin.json 中定义 "plugin.record.created"
    msg := services.I18n().Translate(ctx, "plugin.record.created", "记录创建成功")

    // 发布通知时使用翻译后的文本
    return services.Notify().SendNoticePublication(ctx, &contract.NoticePublication{
        SourceType: "plugin",
        Title:      msg,
        Content:    fmt.Sprintf("记录ID: %d", req.ID),
    })
}
```

#### 动态插件

动态插件通过`guest`侧`SDK`翻译，需要在`plugin.yaml`中声明`service: i18n`：

```go
func CreateRecord(ctx context.Context, req *CreateRecordReq) error {
    msg := guestsdk.I18n.Translate(ctx, "plugin.record.created", "记录创建成功")

    // 发布通知
    return guestsdk.Notify.SendNoticePublication(ctx, &notify.Publication{
        SourceType: "plugin",
        Title:      msg,
        Content:    fmt.Sprintf("记录ID: %d", req.ID),
    })
}
```

### 发布通知

#### 源码插件

源码插件通过`services.Notify()`发布通知：

```go
func (s *recordService) NotifyRecordCreated(ctx context.Context, record *Record) error {
    services := s.services

    title := services.I18n().Translate(ctx, "plugin.record.created.title", "新记录")
    content := services.I18n().Translate(
        ctx,
        "plugin.record.created.content",
        fmt.Sprintf("记录 \"%s\" 已创建", record.Title),
    )

    return services.Notify().SendNoticePublication(ctx, &contract.NoticePublication{
        SourceType: "plugin",
        CategoryCode: "record_created",
        Title:        title,
        Content:      content,
    })
}
```

#### 动态插件

动态插件通过`guest`侧`SDK`发布通知，需要在`plugin.yaml`中声明`service: notify`：

```go
func NotifyRecordCreated(ctx context.Context, record *Record) error {
    title := guestsdk.I18n.Translate(ctx, "plugin.record.created.title", "新记录")
    content := guestsdk.I18n.Translate(
        ctx,
        "plugin.record.created.content",
        fmt.Sprintf("记录 \"%s\" 已创建", record.Title),
    )

    return guestsdk.Notify.SendNoticePublication(ctx, &notify.Publication{
        SourceType:   "plugin",
        CategoryCode: "record_created",
        Title:        title,
        Content:      content,
    })
}
```

### 租户感知

#### 源码插件

源码插件通过`TenantFilterService`实现租户感知的数据操作。该服务在`pluginhost.Services`中扩展提供，不在`capability.Services`中。DAO层由`make dao`生成，提供`dao.Xxx`单例和`do.Xxx`、`entity.Xxx`模型类型。

**服务层注入`TenantFilterService`：**

```go
type serviceImpl struct {
    tenantFilter plugincontract.TenantFilterService
}

func New(tenantFilter plugincontract.TenantFilterService) Service {
    return &serviceImpl{tenantFilter: tenantFilter}
}
```

在插件注册入口中，通过`registrar.Services()`获取并注入：

```go
services := registrar.Services()
recordSvc := recordsvc.New(services.TenantFilter())
```

**查询列表**——使用`Apply()`追加`tenant_id`条件：

```go
func (s *serviceImpl) List(ctx context.Context, pageNum, pageSize int) ([]*entitymodel.Record, int, error) {
    model := s.tenantFilter.Apply(ctx, dao.Record.Ctx(ctx), "")

    total, err := model.Count()
    if err != nil {
        return nil, 0, err
    }

    var items []*entitymodel.Record
    err = model.OrderDesc(dao.Record.Columns().UpdatedAt).
        Page(pageNum, pageSize).
        Scan(&items)
    if err != nil {
        return nil, 0, err
    }

    return items, total, nil
}
```

联合查询时使用`qualifier`参数为`tenant_id`列添加表限定符，避免列名歧义：

```go
func (s *serviceImpl) ListWithAuthor(ctx context.Context) ([]*RecordWithAuthor, error) {
    model := dao.Record.Ctx(ctx).
        LeftJoin("sys_user u", "u.id = "+dao.Record.Columns().AuthorId)

    // qualifier 为 "r"，生成 r.tenant_id 条件
    model = s.tenantFilter.Apply(ctx, model, "r")

    var items []*RecordWithAuthor
    err := model.Scan(&items)
    return items, err
}
```

**新增记录**——通过`Context()`获取租户ID并显式写入：

```go
func (s *serviceImpl) Create(ctx context.Context, in *CreateRecordInput) (int64, error) {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    recordID, err := dao.Record.Ctx(ctx).Data(do.Record{
        TenantId: tenantID,
        Title:    strings.TrimSpace(in.Title),
        Content:  in.Content,
    }).InsertAndGetId()
    if err != nil {
        return 0, err
    }

    return recordID, nil
}
```

**更新和删除**——手动添加`TenantFilterColumn`条件，与主键一起限定范围：

```go
func (s *serviceImpl) Update(ctx context.Context, in *UpdateRecordInput) error {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    _, err := dao.Record.Ctx(ctx).
        Where(plugincontract.TenantFilterColumn, tenantID).
        Where(do.Record{Id: in.Id}).
        Data(do.Record{
            Title:   strings.TrimSpace(in.Title),
            Content: in.Content,
        }).
        Update()
    return err
}

func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
    tenantID := s.tenantFilter.Context(ctx).TenantID

    _, err := dao.Record.Ctx(ctx).
        Where(plugincontract.TenantFilterColumn, tenantID).
        Where(do.Record{Id: id}).
        Delete()
    return err
}
```

#### 动态插件

动态插件有些不一样，由于设计沙箱的权限控制，动态插件需要通过`hostServices`的`data`授权进行数据库查询，租户过滤由主框架在数据访问层自动注入：

```yaml
# plugin.yaml
hostServices:
  - service: data
    methods: [list, get, create, update, delete]
    resources:
      tables:
        - plugin_demo_dynamic_record
```

动态插件中的数据库操作示例：

```go
type recordDAO struct{}

func (d *recordDAO) ListByTenant(ctx context.Context, page, pageSize int) ([]*Record, error) {
    result, err := guestsdk.Data.List(ctx, &data.ListRequest{
        Table:    "plugin_demo_dynamic_record",
        Page:     page,
        PageSize: pageSize,
    })
    if err != nil {
        return nil, err
    }

    var records []*Record
    if err := result.Scan(&records); err != nil {
        return nil, err
    }
    return records, nil
}
```

## 相关内容

import DocCardList from '@theme/DocCardList';

<DocCardList />
