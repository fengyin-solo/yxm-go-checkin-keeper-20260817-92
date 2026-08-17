# BUG_REPRO

## Bug 是什么

奖励流水查询链路中类型过滤、存储返回、排序和 HTTP 参数映射不一致，导致 bonus 流水筛选错误并破坏倒序列表。

## 如何触发

运行目标复现测试：

```bash
go test ./internal/service -run TestRewardGrantFilterAndOrderRegression -count=1
```

## 错误信息

```text
reward_grant_diagnosis_test.go:29: bonus filter should return exactly one grant, total=3 grants=[...]
```
