# Example Clock Go Bindings

This repository provides Go bindings for the Example Clock library, enabling seamless integration with Go projects.

## Installation

To build the required dependencies for this module, the `make` command needs to be executed. If you are integrating this module into another project via `go get`, ensure that you navigate to the `go-bindings-template` module directory and run `make`.

### Steps to Install

Follow these steps to install and set up the module:

1. Retrieve the module using `go get`:
   ```
   go get -u github.com/status-im/go-bindings-template
   ```
2. Navigate to the module's directory:
   ```
   cd $(go list -m -f '{{.Dir}}' github.com/status-im/go-bindings-template)
   ```
3. Prepare third_party directory and clone the example library
   ```
   sudo mkdir third_party
   sudo chown $USER third_party
   ```
4. Build the dependencies:
   ```
   make -C clock
   ```

### How to test

in the `go-bindings-template` directory, please run

```
go test -v ./...
```

Now the module is ready for use in your project.

### Note

In order to easily build the example libclock library on demand, it is recommended to add the following target in your project's Makefile:

```
LIBCLOCK_DEP_PATH=$(shell go list -m -f '{{.Dir}}' github.com/status-im/go-bindings-template)

buildlib:
   cd $(LIBCLOCK_DEP_PATH) &&\
   sudo mkdir -p third_party &&\
   sudo chown $(USER) third_party &&\
   make -C clock
```

## Example Usage

For an example on how to import and use a package, please take a look at our [example-go-bindings](https://github.com/gabrielmer/example-waku-go-bindings) repo
