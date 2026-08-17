# BUG_REPRO

## Bug 是什么

用户局部更新路径破坏了“空字段表示不更新”的约定，状态更新和昵称更新互相覆盖已有字段。

## 如何触发

运行目标复现测试：

```bash
go test ./internal/service -run TestUserUpdatePreservesExistingFields -count=1
```

## 错误信息

```text
user_update_diagnosis_test.go:21: status-only update should preserve nickname, got ""
```
