# Example Clock Go Bindings

This repository provides Go bindings for the [Example Clock library](https://github.com/logos-co/nim-library-template), enabling seamless integration with Go projects.

You can find instructions on how to adapt each file to create Go bindings for your Nim library. All the logic is on `clock.go`

For an example on how it looks on how to integrate the module in other Go projects, please refer to [waku-go-bindings](https://github.com/waku-org/waku-go-bindings)

### How to build

Build the dependencies by running in the root directory of the repository:

```
make -C clock
```

### How to test

in the root directory of the repository, please run

```
go test -v ./...
```
