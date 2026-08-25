## ADDED Requirements

### Requirement: 通用会话存储不得承载带范围列表

系统 SHALL 把会话热状态存储的核心契约限制为写入、读取、删除和续期。带数据权限的在线用户分页列表 MUST 由会话管理投影提供，MUST NOT 作为通用`Store`的核心方法混在认证热路径里。

#### Scenario: 认证路径只依赖核心存储

- **WHEN** 登录、续期或退出调用会话存储
- **THEN** 调用方只使用 Set/Get/Delete/Touch 这类核心方法
- **AND** 不必依赖带`datascope`的列表方法才能编译认证路径
