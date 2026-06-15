## ADDED Requirements

### Requirement: FilesStorage文档说明文件和对象存储边界
文档`7250-files-storage.md` SHALL 说明`Files()`面向宿主治理文件投影，动态`storage`服务面向插件作用域对象存储，两者不应混用。

#### Scenario: 文件入口可区分
- **WHEN** 插件开发者阅读文件与存储能力文档
- **THEN** 理解`Files()`、`Admin().Files()`、`storage.put`、`storage.get`、`storage.delete`、`storage.list`和`storage.stat`的边界

### Requirement: FilesStorage文档说明动态storage资源授权
文档 SHALL 说明动态`storage`服务使用`paths`授权逻辑路径前缀，不暴露宿主物理路径或底层存储后端。

#### Scenario: storage授权边界清晰
- **WHEN** 动态插件开发者声明`service: storage`
- **THEN** 文档要求其声明授权`paths`并避免路径逃逸
