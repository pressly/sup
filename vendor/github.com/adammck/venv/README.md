# venv

[![GoDoc](https://godoc.org/github.com/adammck/venv?status.svg)](https://godoc.org/github.com/adammck/venv)
[![Build Status](https://travis-ci.org/adammck/venv.svg?branch=master)](https://travis-ci.org/adammck/venv)

This is a Go library to abstract access to environment variables.  
Like [spf13/afero][afero] or [blang/vfs][vfs], but for the env.

## Usage

```go
package main

import (
	"fmt"
	"github.com/adammck/venv"
)

func main() {
	var e venv.Env

	// Use the real environment

	e = venv.OS()
	fmt.Printf("Hello, %s!\n", e.Getenv("USER"))

	// Or use a mock

	e = venv.Mock()
	e.Setenv("USER", "fred")
	fmt.Printf("Hello, %s!\n", e.Getenv("USER"))
}
```

## License

MIT.

[afero]: https://github.com/spf13/afero
[vfs]: https://github.com/blang/vfs
