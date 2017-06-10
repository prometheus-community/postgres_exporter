
Vendored versions of the build tooling.

gocovmerge is used to merge coverage reports for uploading to a service like
coveralls, and gometalinter conveniently incorporates multiple Go linters.

By vendoring both, we gain a self-contained build system.

Run `make all` to build, and `make update` to update.
