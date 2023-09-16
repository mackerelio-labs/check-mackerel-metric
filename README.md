# check-mackerel-metric

## Description

Checks Mackerel host metrics are still being posted.

It is also available in the host metric of cloud integrations.

## Synopsis
```
check-mackerel-metric -H HOST_ID -n METRIC_NAME -w WARNING_MINUTE -c CRITICAL_MINUTE
```

CRITICAL (or WARNING) alert is issued if no metric has been posted since the minute specified for CRITICAL_MINUTE (or WARNING_MINUTE) from the current time.

## Setting for mackerel-agent
```
[plugin.checks.metric-myhost]
command = ["check-mackerel-metric", "-H", "HOST_ID", "-n", "METRIC_NAME", "-w", "WARNING_MINUTE", "-c", "CRITICAL_MINUTE"]
```

## Usage
### Options
```
-H, --host=     target host ID
-n, --name=     target metric name
-w, --warning=  minute to be WARNING
-c, --critical= minute to be CRITICAL
```

- HOST_ID is displayed at the top of the Mackerel host screen, like `4Hkc5RWzXXX`.
- METRIC_NAME can be looked up with `mkr metric-names -H HOST_ID`.

--

## 説明

Mackerelのホストメトリックの投稿が継続していることをチェックします。

クラウドインテグレーションのホストメトリックでも利用可能です。

## 概要
```
check-mackerel-metric -H HOST_ID -n METRIC_NAME -w WARNING_MINUTE -c CRITICAL_MINUTE
```

現在時刻から CRITICAL_MINUTE (または WARNING_MINUTE) 分前の間に何もメトリックの投稿がないときに、CRITICAL (または WARNING) アラートが発報されます。

## mackerel-agentでの設定
```
[plugin.checks.metric-myhost]
command = ["check-mackerel-metric", "-H", "HOST_ID", "-n", "METRIC_NAME", "-w", "WARNING_MINUTE", "-c", "CRITICAL_MINUTE"]
```

## 使い方
### オプション
```
-H, --host=     対象のホストID
-n, --name=     対象のメトリック名
-w, --warning=  指定の分数内にメトリックがなければWARNING
-c, --critical= 指定の分数内にメトリックがなければCRITICAL
```

- HOST_ID (ホストID) はMackerelのホスト画面の上部に `4Hkc5RWzXXX` のように表示されています。
- METRIC_NAME (メトリック名) は `mkr metric-names -H HOST_ID` で調べることができます。
