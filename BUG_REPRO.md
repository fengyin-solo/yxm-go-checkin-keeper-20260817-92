# BUG_REPRO

## Bug 是什么

签到记录日期区间过滤、列表排序、月历聚合和最近连续天数统计的跨层契约被破坏，导致同一批签到记录在列表与日历接口里表现不一致。

## 如何触发

运行目标回归测试：

```bash
go test ./internal/service -run TestCalendarAndDateRangeRegression -count=20
```

## 错误信息

```text
calendar_filter_regression_test.go:29: date range should include exactly 3 records, got 0
```
