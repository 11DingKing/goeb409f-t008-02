# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

值班室报了个很离谱的现象，麻烦先帮我们定位原因，暂时不要改代码。

现象：主网刚失电、值长第一时间下达真实黑启动令，如果调度台没填「主网失电时刻」这个可选字段，
过程一创建出来就被标成「黑启动超时」了。

复现：
1. POST /api/v1/processes  body 只有 {"kind":"real"}（不带 grid_lost_at）
2. 返回的 JSON 里 "command_overdue":true
3. GET /api/v1/processes/{id} 看事件流，除了 black_start_commanded 还多出一条 black_start_overdue，
   detail 是 "command issued 2562047h47m16.854775807s after the 5m0s window"
4. 同一时刻带上 grid_lost_at（例如 5 小时前）下令，command_overdue 也是 true —— 这个是对的，确实超了 5 分钟窗口
5. 带上 1 分钟前的 grid_lost_at 下令，command_overdue 是 false —— 这个也是对的

所以问题只出在「没填失电时刻」这条路径上。业务上这个字段是可选的：现场有时候确实来不及记准失电时刻，
这时候我们期望的是不判超时（没有基准就没法判），而不是直接扣一个超时帽子。
现在超时事件会进考核台账，值长被无故记了一次黑启动超时。那个 2562047 小时的数字也很可疑。

请先不要修改任何代码，只做定位。我们需要：
- 出问题的具体 Go 文件和具体符号
- 该符号的什么错误行为造成的
- 它为什么会让「没填失电时刻」的真实黑启动令立刻被判超时，以及那个 2562047 小时是怎么来的（完整因果机制）
- 你自己实际跑出来的证据（执行了什么命令、看到什么输出）
临时复现程序请放在仓库之外的临时目录，不要改动仓库里的文件。

## 含 Bug 版本

- 仓库：11DingKing/goeb409f-t008-02
- 仓库地址：https://github.com/11DingKing/goeb409f-t008-02.git
- parent SHA：10921e32dfaca590f05c629fe19a4eb8efec29af

## 复现步骤

```bash
git clone -- https://github.com/11DingKing/goeb409f-t008-02.git bug-repro
cd bug-repro
git checkout --detach 10921e32dfaca590f05c629fe19a4eb8efec29af
go test ./internal/domain/ -run "^TestProcess_CommandWindowWhenGridLossTimeUnknown$|^TestProcess_CommandOverdue$" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/domain/ -run "^TestProcess_CommandWindowWhenGridLossTimeUnknown$|^TestProcess_CommandOverdue$" -count=1 -v
FAIL	microgrid/internal/domain [build failed]
FAIL

```

stderr：

```text
# microgrid/internal/domain [microgrid/internal/domain.test]
internal/domain/command_window_test.go:14:9: undefined: fixedTime
internal/domain/command_window_test.go:20:5: undefined: hasEvent
internal/domain/command_window_test.go:39:6: undefined: hasEvent

```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/domain/ -run "^TestProcess_CommandWindowWhenGridLossTimeUnknown$|^TestProcess_CommandOverdue$" -count=1 -v
FAIL	microgrid/internal/domain [build failed]
FAIL

```

stderr：

```text
# microgrid/internal/domain [microgrid/internal/domain.test]
internal/domain/command_window_test.go:14:9: undefined: fixedTime
internal/domain/command_window_test.go:20:5: undefined: hasEvent
internal/domain/command_window_test.go:39:6: undefined: hasEvent

```

## 通过条件

目标仓库工作区零改动：git status --porcelain 为空，执行前后 tree hash 一致，生产代码、测试与配置均未被修改。
指出具体 Go 文件与具体符号，并说明该符号的错误行为如何导致题面症状，因果机制完整（含那个异常大的时长是怎么产生的）。
给出自己实际运行得到的证据（命令与输出），不能只做静态推断。
结论需与 gold_root_cause 的文件、符号和失效机制一致。
允许在仓库之外的临时目录写一次性复现程序；不产生代码修复提交。
