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
		fmt.Fprintln(os.Stdout, "dlm: need 1 argument")
		os.Exit(2)
	}
	err := run(os.Args[1], os.Getenv("HOME")+"/Downloads")
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
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
	if err := download(rawurl, dir); err != nil {
		return err
	}
	return nil
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
	_, err = io.CopyBuffer(w, resp.Body, make([]byte, 1<<20))
	if err != nil {
		return err
	}
	fmt.Fprintf(w.output, "\r%3.f%% %[4]*[2]d/%d", 100*float32(w.n)/float32(w.l), w.n, w.l, len(strconv.Itoa(w.l)))
	w.output.Write([]byte("\n"))
	return nil
}

func newWriter(w, output io.Writer, l int) *writer {
	nw := &writer{
		w:      w,
		output: output,
		l:      l,
	}
	go nw.log()
	return nw
}

type writer struct {
	w      io.Writer
	output io.Writer
	l      int

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

func (w *writer) log() {
	c := time.Tick(100 * time.Millisecond)
	l := w.l
	for range c {
		w.mu.Lock()
		n := w.n
		w.mu.Unlock()
		fmt.Fprintf(w.output, "\r%3.f%% %[4]*[2]d/%d", 100*float32(n)/float32(l), n, l, len(strconv.Itoa(l)))
	}
}
