

## GAS - Go AST Scanner

Inspects source code for security problems by scanning the Go AST.

### License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License [here](http://www.apache.org/licenses/LICENSE-2.0).

### Project status

[![Build Status](https://travis-ci.org/GoASTScanner/gas.svg?branch=master)](https://travis-ci.org/GoASTScanner/gas)
[![GoDoc](https://godoc.org/github.com/GoASTScanner/gas?status.svg)](https://godoc.org/github.com/GoASTScanner/gas)

Gas is still in alpha and accepting feedback from early adopters. We do
not consider it production ready at this time.

### Usage

Gas can be configured to only run a subset of rules, to exclude certain file
paths, and produce reports in different formats. By default all rules will be
run against the supplied input files. To recursively scan from the current
directory you can supply './...' as the input argument.

#### Selecting rules

By default Gas will run all rules against the supplied file paths. It is however possible to select a subset of rules to run via the '-include=' flag,
or to specify a set of rules to explicitly exclude using the '-exclude=' flag.

##### Available rules

  - G101: Look for hardcoded credentials
  - G102: Bind to all interfaces
  - G103: Audit the use of unsafe block
  - G104: Audit errors not checked
  - G105: Audit the use of math/big.Int.Exp
  - G201: SQL query construction using format string
  - G202: SQL query construction using string concatenation
  - G203: Use of unescaped data in HTML templates
  - G204: Audit use of command execution
  - G301: Poor file permissions used when creating a directory
  - G302: Poor file permisions used with chmod
  - G303: Creating tempfile using a predictable path
  - G401: Detect the usage of DES, RC4, or MD5
  - G402: Look for bad TLS connection settings
  - G403: Ensure minimum RSA key length of 2048 bits
  - G404: Insecure random number source (rand)
  - G501: Import blacklist: crypto/md5
  - G502: Import blacklist: crypto/des
  - G503: Import blacklist: crypto/rc4
  - G504: Import blacklist: net/http/cgi


```
# Run a specific set of rules
$ gas -include=G101,G203,G401 ./...

# Run everything except for rule G303
$ gas -exclude=G303 ./...
```

#### Excluding files:

Gas can be told to \ignore paths that match a supplied pattern using the 'skip' command line option. This is
accomplished via [go-glob](github.com/ryanuber/go-glob). Multiple patterns can be specified as follows:

```
$ gas -skip=tests* -skip=*_example.go ./...
```

#### Annotating code

As with all automated detection tools there will be cases of false positives. In cases where Gas reports a failure that has been manually verified as being safe it is possible to annotate the code with a '#nosec' comment.

The annotation causes Gas to stop processing any further nodes within the
AST so can apply to a whole block or more granularly to a single expression.

```go

import "md5" // #nosec


func main(){

    /* #nosec */
    if x > y {
        h := md5.New() // this will also be ignored
    }

}

```

In some cases you may also want to revisit places where #nosec annotations
have been used. To run the scanner and ignore any #nosec annotations you
can do the following:

```
$ gas -nosec=true ./...
```

### Output formats

Gas currently supports text, json and csv output formats. By default
results will be reported to stdout, but can also be written to an output
file. The output format is controlled by the '-fmt' flag, and the output file is controlled by the '-out' flag as follows:

```
# Write output in json format to results.json
$ gas -fmt=json -out=results.json *.go
```

### Docker container

A Dockerfile is included with the Gas source code to provide a container that 
allows users to easily run Gas on their code. It builds Gas, then runs it on 
all Go files in your current directory. Use the following commands to build 
and run locally:

To build: (run command in cloned Gas source code directory)
          docker build --build-arg http_proxy --build-arg https_proxy
          --build-arg no_proxy -t goastscanner/gas:latest .

To run:  (run command in desired directory with Go files)
          docker run -v $PWD:$PWD --workdir $PWD goastscanner/gas:latest

Note: Docker version 17.05 or later is required (to permit multistage build).
```
