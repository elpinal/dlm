package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
)

const defaultProg = "curl"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stdout, "dlm: need 1 argument")
		os.Exit(2)
	}
	err := run(defaultProg, os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func run(prog string, rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	dir := u.Host + "/" + path.Dir(u.Path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	if err := download(prog, rawurl, dir); err != nil {
		return err
	}
	return nil
}

func download(prog string, url, dir string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, "-O", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
