# Using Go

## Import paths and `$GOPATH`

Go organises code in packages. A package is basically a directory containing Go source files.

Go looks for and stores its packages in the path(s) pointed to by the `$GOPATH` environment variable. Before you do anything related to Go, you ought to set this variable in your shell's start-up file, such as `.bashrc` for the Bash shell or `.zshrc` for Zsh. If your system uses PAM – most Linux distributions do – then you can also set the variable in `~/.pam_environment` so it will be set for any shell:

```
GOPATH DEFAULT="@{HOME}/Code/Go"
GOBIN DEFAULT="@{HOME}/.local/bin"
```

This will set the path to `~/.Code/Go` and store installed binaries in `~/.local/bin`.

If you want to keep your own code separate from (automatically) downloaded packages, you can set `$GOPATH` to include two paths, similar to the POSIX system `$PATH` variable:

```
export GOPATH="~/.cache/go:~/gocode"
```

Go will download package dependencies by default in the first path `~/.cache/go` and also search `~/gocode` for your own packages.

When you import a package into your own code, Go will search its source directory `$GOROOT` and all the paths in `$GOPATH` for that package:

```go
package main

import (
	"fmt"
	
	"github.com/yourgithubaccount/helloworld"
)

func main() {
	msg := helloworld.Greet()
	fmt.Println(msg)
}

```

During compilation, Go will look for a directory `github.com/yourgithubaccount/helloworld` in `$GOROOT/src` and `$GOPATH/src` for all paths in `$GOPATH`. This is also where `go get` will download the package(s) to when you pull in dependencies.

When you import a package with import path IMPORT_PATH, Go will use the following paths by default:

- `$GOPATH/src/IMPORT_PATH` for the source;
- `$GOPATH/pkg/OS_ARCH/IMPORT_PATH` for compiled libraries;
- `$GOPATH/bin` for any installed commands (packages with a `main` function).

You can check used environment variables with the `go env` command.

In short, any code not in `$GOPATH` can not be referred to – at least for current versions of Go – so set `$GOPATH` and keep your own code below that to make sure packages can find and import each other.


## Code organisation

As Go packages are directories, Go expects to find only one package in a directory. If you want to build different programs sharing similar code, it is customary to add a directory called `cmd/` with below that package directories for each command. The other directories than refer to library packages.

The difference between a command and a library is the existence of a `main()` function in the former; as libraries are called by commands, they have no use for a `main()` function.

So if you were to create a program with two libraries and two commands, you would create a directory – let's say `myproject` – under `$GOPATH/src`.

In `$GOPATH/src/myproject`, you would then have:

- `dosomething/dosomething.go`, with package declaration: `package dosomething`;
- `smtelse/smtelse.go`, with package declaration: `package smtelse`;
- `cmd/smt/main.go`, with package declaration: `package main`, and import statements:
  - `myproject/dosomething`
  - `myproject/smtelse`
- cmd/doit/main.go, with package declaration: `package main`, and imports:
  - `myproject/smtelse`

You can now build the binary in `$GOPATH/src/myproject/cmd/smt` with `go build` and it will compile in the two library packages you imported.

If you share your code on for instance Github, you wouldn't use `$GOPATH/src/myproject` but rather `$GOPATH/src/github.com/YOURUSERNAME/myproject` as the base directory. This would make your import path for the `dosomething` library `github.com/YOURUSERNAME/myproject/dosomething`, which would make your code "go get"-able, meaning others could both download your packages and use existing import paths in them without change. When people try to use your program, `go get` would by default put the packages in their `$GOPATH` –  such as `$GOPATH/src/github.com/YOURUSERNAME/myproject` – and the import paths would continue to work both relative to each other and absolute in regard to the download location. Even though people could have different locations for `$GOPATH`, the structure below that path would be the same for everyone.

Anybody could then get your whole project with a simple:

```shell
go get github.com/YOURUSERNAME/myproject
```

-or- import its library packages in their own code:

```go
import (
	"github.com/YOURUSERNAME/myproject/dosomething"
	"github.com/YOURUSERNAME/myproject/smtelse"
)
```

The way Go has chosen to solve imports and dependencies means you give up some flexibility and get longer import paths, but you gain simplicity and transparency, knowing it should work the same for everyone.


## See also

- [How to Write Go Code](https://golang.org/doc/code.html) from the offical Go documentation
- [About the go command](https://golang.org/doc/code.html) from the official Go documentation
- [Organizing Go code](https://blog.golang.org/organizing-go-code) from the official Go Blog
- [Godoc: documenting Go code](https://blog.golang.org/godoc-documenting-go-code) from the official Go blog
- [Structuring Applications in Go](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091) by Ben Johnson
