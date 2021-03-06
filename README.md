# Dlm

`dlm` is a download manager.

## Install

To install, use `go get`:

```bash
$ go get -u github.com/elpinal/dlm
```

-u flag stands for "update".

## Examples

To download files:

```bash
$ dlm example.com golang.org/pkg/path/filepath golang.org/ref/spec
1270 bytes
100% 28339/28339 bytes
100% 219533/219533 bytes
```

Then `open` the local content.

```bash
$ dlm -open example.com
```

These files are stored in the `$HOME/Downloads` folder.
Only macOS is supported.

## Contribution

1. Fork ([https://github.com/elpinal/dlm/fork](https://github.com/elpinal/dlm/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[elpinal](https://github.com/elpinal)
