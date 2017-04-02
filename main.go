package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "dlm: need 1 argument")
		os.Exit(2)
	}
	err := run(os.Args[1], os.Getenv("HOME")+"/Downloads")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(rawurl string, prefix string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	dir := path.Join(prefix, "/", u.Host, "/", path.Dir(u.Path))
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	return download(rawurl, dir)
}

func download(url, dir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	l, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, path.Base(url)))
	if err != nil {
		return err
	}
	defer f.Close()
	w := newWriter(f, os.Stdout, l)
	bufSize := 1 << 20
	if bufSize > l {
		bufSize = l
	}
	_, err = io.CopyBuffer(w, resp.Body, make([]byte, bufSize))
	if err != nil {
		return err
	}
	w.log()
	w.output.Write([]byte("\n"))
	return nil
}

func newWriter(w, output io.Writer, l int) *writer {
	nw := &writer{
		w:      w,
		output: output,
		l:      l,
		width:  len(strconv.Itoa(l)),
	}
	go nw.interval()
	return nw
}

type writer struct {
	w      io.Writer
	output io.Writer

	l     int
	width int

	mu sync.Mutex
	n  int
}

func (w *writer) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		return 0, err
	}
	w.mu.Lock()
	w.n += n
	w.mu.Unlock()
	return n, nil
}

func (w *writer) interval() {
	c := time.Tick(100 * time.Millisecond)
	for range c {
		w.log()
	}
}

func (w *writer) log() {
	w.mu.Lock()
	n := w.n
	w.mu.Unlock()
	fmt.Fprintf(w.output, "\r%3d%% %[4]*[2]d/%d", w.percentage(n), n, w.l, w.width)
}

func (w *writer) percentage(n int) int {
	return 100 * n / w.l
}
