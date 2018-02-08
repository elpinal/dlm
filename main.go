package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Dlm is a download manager")
		fmt.Fprintln(os.Stderr, "Usage: dlm [flags] urls...")
		flag.PrintDefaults()
	}
	flagOpen := flag.Bool("open", false, "open downloaded content")
	flagGzip := flag.Bool("gzip", false, "decompress gzip files")
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "dlm: need 1 or more arguments")
		os.Exit(2)
	}
	if runtime.GOOS != "darwin" {
		fmt.Fprintf(os.Stderr, "dlm: warning: on %s, it may not work well\n", runtime.GOOS)
	}
	prefix := filepath.Join(os.Getenv("HOME"), "Downloads")
	for _, arg := range flag.Args() {
		err := run(arg, prefix, *flagOpen, *flagGzip)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func run(rawurl string, prefix string, flagOpen bool, flagGzip bool) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	dir := filepath.Join(prefix, u.Host, path.Dir(u.Path))
	if flagOpen {
		return open(rawurl, dir)
	}
	if flagGzip {
		return gzip(rawurl, dir)
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	return download(rawurl, dir)
}

func open(url string, dir string) error {
	cmd := exec.Command("open", filepath.Join(dir, path.Base(url)))
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gzip(url string, dir string) error {
	cmd := exec.Command("gzip", "--decompress", filepath.Join(dir, path.Base(url)))
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func download(url, dir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Ignore error
	l, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	f, err := os.Create(filepath.Join(dir, path.Base(url)))
	if err != nil {
		return err
	}
	defer f.Close()
	w := newWriter(f, os.Stdout, l)
	bufSize := 1 << 20
	if l > 0 && bufSize > l {
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
	if w == nil {
		w = ioutil.Discard
	}
	if output == nil {
		output = ioutil.Discard
	}
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
	output io.Writer // for logging

	// read-only
	l     int // content length
	width int // width of "content length as string"

	mu sync.Mutex
	n  int // read count
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

// log prints the percentage of read bytes, read bytes and content bytes.
func (w *writer) log() {
	w.mu.Lock()
	n := w.n
	w.mu.Unlock()
	if w.l == 0 {
		fmt.Fprintf(w.output, "\r%d bytes", n)
		return
	}
	fmt.Fprintf(w.output, "\r%3d%% %[4]*[2]d/%d bytes", w.percentage(n), n, w.l, w.width)
}

func (w *writer) percentage(n int) int {
	return 100 * n / w.l
}
