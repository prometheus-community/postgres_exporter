# unparam

[![Build Status](https://travis-ci.org/mvdan/unparam.svg?branch=master)](https://travis-ci.org/mvdan/unparam)

	go get -u github.com/mvdan/unparam

Reports unused function parameters in your code.

To minimise false positives, it ignores:

* Unnamed and underscore parameters
* Funcs whose signature matches a reachable func type
* Funcs whose signature matches a reachable interface method
* Funcs that have empty bodies
* Funcs that will almost immediately panic or return constants

False positives can still occur by design. The aim of the tool is to be
as precise as possible - if you find any, file a bug.

Note that "reachable" means func signatures found in top-level
declarations in each package and all of its direct dependencies. The
tool ignores transitive dependencies and local signatures.
