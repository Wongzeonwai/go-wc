package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// workcount
type MyWc struct {
	cmds    map[Cmd]bool
	readers []*rAndErr
	isStdin bool // 是否标准输入
}

// 定义输入命令的类型
type Cmd string

var (
	IsByte Cmd = "bytes" // -c
	IsLine Cmd = "lines" // -l
	IsWord Cmd = "words" // -w
	IsChar Cmd = "chars" // -m
)

// 数据源和文件打开的错误
type rAndErr struct {
	r   io.ReadSeeker
	err error
}

func NewWc() (wc *MyWc, close func()) {
	// 获取命令行输入（输入，默认值，用法）
	c := flag.Bool("c", false, "print the byte counts")
	l := flag.Bool("l", false, "print the newline counts")
	w := flag.Bool("w", false, "print the word counts")
	m := flag.Bool("m", false, "print the character counts")

	flag.Parse()

	cmds := make(map[Cmd]bool)

	cmds[IsByte] = *c
	cmds[IsLine] = *l
	cmds[IsWord] = *w
	cmds[IsChar] = *m

	fs := flag.Args() // 文件名切片

	wc = &MyWc{
		cmds: cmds,
	}

	readers := make([]*rAndErr, 0, len(fs))

	for _, fname := range fs {
		f, err := os.Open(fname)

		fe := &rAndErr{
			r:   f,
			err: err,
		}

		readers = append(readers, fe)
	}

	if len(readers) == 0 {
		input, err := io.ReadAll(os.Stdin)
		r := bytes.NewReader(input)
		fe := &rAndErr{
			r:   r,
			err: err,
		}
		readers = append(readers, fe)
		wc.isStdin = true
	}

	wc.readers = readers

	// 关闭文件资源
	close = func() {
		for _, re := range wc.readers {
			if re.err == nil {
				if c, ok := re.r.(io.Closer); ok {
					c.Close()
				}
			}
		}
	}

	return wc, close
}

func (wc *MyWc) WriteTo(w io.Writer) (int, error) {
	keys := make([]Cmd, 0, len(wc.cmds))
	for cmd, ok := range wc.cmds {
		if ok {
			keys = append(keys, cmd)
		}
	}
	slices.Sort(keys)

	// 设置打印表头
	header := make([]string, len(keys))

	for i, v := range keys {
		header[i] = string(v)
	}

	if !wc.isStdin {
		header = append(header, "file")
	}

	results := make([][]string, 0, len(wc.readers)+1)

	results = append(results, header)

	// 确定每一列的最小宽度，确保打印齐整
	minWidth := make([]int, len(header))

	for i, v := range header {
		minWidth[i] = len(v)
	}

	for _, re := range wc.readers {
		if re.err != nil {
			results = append(results, []string{re.err.Error()})
			continue
		}

		counts := wc.Count(re.r, keys)
		result := make([]string, 0, len(counts)+1)

		for i, c := range counts {
			str := strconv.Itoa(c)
			if len(str) > minWidth[i] {
				minWidth[i] = len(str)
			}
			result = append(result, str)
		}

		if !wc.isStdin {
			if f, ok := re.r.(*os.File); ok {
				name := f.Name()
				result = append(result, name)

				if len(name) > minWidth[len(minWidth)-1] {
					minWidth[len(minWidth)-1] = len(name)
				}
			}
		}

		results = append(results, result)
	}

	format := formatResult(results, minWidth)
	return w.Write([]byte(format))
}

func (wc *MyWc) Count(r io.ReadSeeker, keys []Cmd) []int {
	cs := make([]int, 0, len(keys))

	for _, cmd := range keys {
		if !wc.cmds[cmd] {
			continue
		}

		result := count(r, which(cmd))
		cs = append(cs, result)
	}

	return cs
}

func count(r io.ReadSeeker, by bufio.SplitFunc) int {
	scanner := bufio.NewScanner(r)
	scanner.Split(by)

	var total int

	for scanner.Scan() {
		total++
	}

	// 重置文件游标，从头开始读取
	r.Seek(0, io.SeekStart)

	return total
}

func which(by Cmd) bufio.SplitFunc {
	switch by {
	case IsByte:
		return bufio.ScanBytes
	case IsChar:
		return bufio.ScanRunes
	case IsLine:
		return bufio.ScanLines
	default:
		return bufio.ScanLines
	}
}

func formatResult(results [][]string, minWidth []int) string {
	var builder strings.Builder

	for _, row := range results {
		if len(row) == 1 {
			builder.WriteString(row[0])
			builder.WriteString("\n")
			continue
		}

		for i, v := range row {
			if i > 0 {
				builder.WriteString(" ")
			}
			// *表示设置动态宽度，-表示左对齐，默认右对齐
			builder.WriteString(fmt.Sprintf("%-*s", minWidth[i], v))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func (wc *MyWc) SetDefaultCmd() {
	var has bool

	for _, ok := range wc.cmds {
		if ok {
			has = true
		}
	}

	if !has {
		wc.cmds[IsByte] = true
		wc.cmds[IsLine] = true
		wc.cmds[IsWord] = true
	}
}
