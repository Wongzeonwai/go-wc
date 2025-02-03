package main

import "os"

func main() {
	wc, close := NewWc()

	defer close()

	wc.SetDefaultCmd()
	wc.WriteTo(os.Stdout)
}
