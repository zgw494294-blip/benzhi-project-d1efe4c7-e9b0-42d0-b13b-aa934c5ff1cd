# 触导标识核准台

用于公共建筑无障碍触觉导向标识版样的建档、修订实测、确定性合规校核、整改复核、冻结和制作授权验证。服务由 Go 原生提供响应式浏览器工作台与同源 JSON API，数据保存在本地 bbolt 文件。

工作台现支持建档字段预检与持久化幂等保护、标准目录选择、不可变修订差异和规则影响预览、六类实测证据覆盖矩阵、校核运行历史及相邻修订迁移对比、逐项整改闭环、结构化退回项回应与确认、冻结清单根摘要预览确认，以及凭据 JSON 的 `VALID`、`TAMPERED`、`UNKNOWN`、`MISMATCHED` 本地核验记录。

浏览器入口：`GET /`。同源 API 延续 `/api/cases` 与 `/api/cases/{case_id}/{action}` 主路径；标准目录为 `GET /api/standards`。建档必须提供 `idempotencyKey`，冻结确认必须提供预览返回的 `digest` 作为 `previewDigest`。

标准构建：`go build ./...`

运行服务：`go run ./cmd/tactile-review -addr=127.0.0.1:19081`

运行测试：`go test ./...`

端到端自检：`go run ./cmd/tactile-review -selfcheck -addr=127.0.0.1:19081`
