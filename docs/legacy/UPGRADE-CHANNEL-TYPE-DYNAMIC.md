# 动态渠道类型与参数模板机制说明

## 1. ChannelType 结构与注册机制

- 每个渠道类型通过 `channeltype.ChannelTypeInfoV2` 注册，包含：
  - `ID`：类型唯一标识（int）
  - `Name`：英文名（string）
  - `Label`：前端展示名（string）
  - `Category`：分组（string）
  - `Description`：描述（string，可选）
  - `Template`：参数模板（[]ChannelTypeTemplateField）

- 注册方式：
```go
channeltype.RegisterChannelType(channeltype.ChannelTypeInfoV2{
    ID: 1001,
    Name: "mock",
    Label: "Mock类型",
    Category: "test",
    Description: "测试类型",
    Template: channeltype.ChannelTypeTemplate{
        {Name: "region", Type: "string", Required: true, Description: "区域"},
        {Name: "sk", Type: "string", Required: false, Description: "密钥"},
    },
})
```

- 支持热插拔/扩展，所有类型和模板均通过注册表维护。

## 2. 参数模板结构

- 字段定义：
```go
type ChannelTypeTemplateField struct {
    Name        string        // 字段名
    Type        string        // string/number/bool/select
    Required    bool          // 是否必填
    Default     interface{}   // 默认值
    Description string        // 字段描述
    Options     []interface{} // 可选项（select 用）
    Pattern     string        // 校验正则
}
```

- 示例：
```json
[
  {"name": "region", "type": "string", "required": true, "desc": "区域"},
  {"name": "sk", "type": "string", "required": false, "desc": "密钥"}
]
```

## 3. 查询接口

- 路由：`GET /api/channel/types`
- 返回：所有已注册类型及参数模板
- 响应结构：
```json
{
  "success": true,
  "data": [
    {
      "id": 1001,
      "name": "mock",
      "label": "Mock类型",
      "category": "test",
      "description": "测试类型",
      "template": [ ... ]
    },
    ...
  ]
}
```

## 4. 参数校验机制

- 新增/编辑渠道时，后端自动根据模板校验必填字段和格式（pattern）。
- 校验失败时接口返回 200 + `success: false` + message。
- 支持扩展更多校验类型。

## 5. 扩展指引

- 新增类型：实现注册代码，指定模板字段。
- 修改模板：直接调整注册参数。
- 前端自动感知类型和参数，无需手动同步。

---

如需高级用法、二次开发或遇到问题，请参考源码注释或联系维护者。
# UniAPI 渠道类型动态化升级方案（TDD）

本方案将渠道类型（channel type）下拉选项由前端硬编码改为后端接口动态提供，提升灵活性和一致性。采用 TDD（测试驱动开发）方式推进。

---

## 1. 目标
- 新增后端接口 `/api/channel/types`，返回所有支持的渠道类型（id、name、label/desc）。
- 前端 select 渠道类型时动态请求该接口渲染选项。
- 保证升级过程有自动化测试覆盖。

---

## 2. 步骤概览
1. 设计接口返回格式与用例，编写后端测试（controller 层单元测试）。
2. 实现后端接口（controller、router、channeltype）。
3. 前端编写接口 mock 测试与集成测试。
4. 前端 select 组件改为动态加载。
5. 验证端到端流程，完善文档。

---

## 3. 详细步骤

### 3.1 设计接口与测试用例
- 路径：`GET /api/channel/types`
- 返回：`[{ id: number, name: string, label: string }]`
- 用例：
  - 能返回所有渠道类型，且与 channeltype/define.go 保持一致。
  - 返回内容包含 id、name、label 字段。

### 3.2 后端开发
1. 在 `relay/channeltype/define.go` 增加类型与 label 映射（如 map[int]string）。
2. 在 `controller/channel.go` 新增 `GetChannelTypes` 方法，返回所有类型。
3. 在 `router/api.go` 注册 `/api/channel/types` 路由。
4. 编写 controller 层单元测试，覆盖接口返回内容。

### 3.3 前端开发
1. 在 `src/pages/channels` 相关组件添加接口 mock 测试。
2. 修改 select 组件，页面加载时 fetch `/api/channel/types`，渲染选项。
3. 保证前端测试覆盖（如 jest/rtl）。

### 3.4 集成与验收
1. 启动后端、前端，手动测试渠道创建流程。
2. 检查 select 选项与后端接口一致。
3. 补充升级说明与回滚方案。

---

## 4. 回滚方案
- 若升级异常，可将前端 select 组件回退为原有硬编码方式。
- 保留原有渠道类型常量，便于快速切换。

---

## 5. 相关文件
- 后端：`relay/channeltype/define.go`、`controller/channel.go`、`router/api.go`
- 前端：`web/modern/src/pages/channels/components/ChannelBasicInfo.tsx` 及相关 select 组件
- 测试：`controller/channel_types_test.go`、`web/modern/src/pages/channels/__tests__/ChannelBasicInfo.test.tsx`

---

## 6. 备注
- 如需国际化，label 可多语言。
- 若需权限控制，可在接口中按用户角色过滤类型。

---

> 按上述步骤推进，每步先写测试再实现功能，确保升级安全可回滚。
