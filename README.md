# gostd2joker

Idea here is for people who build Joker to optionally run (a future version of) this tool against a Go source tree, which _must_ correspond to the version of Go they use to build Joker itself, to populate the `std/go/` subdirectory under Joker itself.

At the moment, this is just a proof of concept, focusing on `net.LookupMX()`. E.g. run it like this:

```
$ ./gostd2joker --dir ~/github/golang/go/src 2>&1 | less
```

Then page through it. Code snippets intended for e.g. `net.joke` are currently just printed to `stdout`, making iteration much easier.

Anything not supported results in either a `panic` or, more often, the string `ABEND` along with some kind of explanation. The latter is used to auto-detect a non-convertible function, in which case the snippet(s) are still output, but commented-out, so it's easy to see what's missing and (perhaps) why.

Among things to do to "productize" this:

* Add `--populate <dir>` option to specify the `std/go/` directory into which to write actual code
* Generate imports and such properly
* Might have to replace the current ad-hoc tracking of Go packages with something that respects `import` and the like
* Explain how to use the tool in Joker's `README.md`
