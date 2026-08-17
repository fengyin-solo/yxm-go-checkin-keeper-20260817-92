# BUG_REPRO

## Bug 是什么

用户列表的过滤、排序、存储列表返回和 HTTP 查询参数传递出现跨层不一致，导致状态筛选、昵称关键词匹配和分页顺序错误。

## 如何触发

运行目标回归测试：

```bash
go test ./internal/service -run TestUserListFilterSortAndPaginationRegression -count=20
```

## 错误信息

```text
user_list_regression_test.go:27: active filter should return 2 users, total=0 users=[]
```
