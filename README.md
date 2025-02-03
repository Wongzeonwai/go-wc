# go-wc
一个使用Go语言实现的wc命令，主要功能：
- 打印文件行数`-l`
- 打印文件字节数`-c`
- 打印文件字符数`-m`
- 打印文件单词数`-w`

例如：
```shell
./go-wc -l go.mod
```
输出：
```shell
lines file
3  go.mod
```
接收标准输入，例如：
```shell
cat go-wc.go | ./go-wc -l -c -w -m
```
输出：
```shell
bytes chars lines words
4333  4333  247   247
```
