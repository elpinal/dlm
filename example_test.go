package main

import "os"

func ExampleLog() {
	w := newWriter(nil, os.Stdout, 1000)
	w.n = 840
	w.Write(make([]byte, 10))
	w.log()
	// Output:
	//  85%  850/1000 bytes
}
