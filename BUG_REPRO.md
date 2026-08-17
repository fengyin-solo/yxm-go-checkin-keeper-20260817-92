# BUG_REPRO

## Bug 是什么

服务层多个校验失败路径没有稳定返回 ValidationError，奖励规则关联活动、活动非法流转、停用用户签到和规则状态更新的错误契约不一致。

## 如何触发

运行目标回归测试：

```bash
go test ./internal/service -run TestServiceValidationErrorsStayTyped -count=20
```

## 错误信息

```text
error_contract_regression_test.go:16: missing activity on reward rule should be a validation error, got 记录不存在
```
