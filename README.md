# gostd2joker

Idea here is for people who build Joker to optionally run (a future version of) this tool against a Go source tree, which _must_ correspond to the version of Go they use to build Joker itself, to populate the `std/go/` subdirectory under Joker itself. Further, the build parameters (`$GOOS`, `$GOARCH`, etc.) must match -- so `build-all.sh` would have to pass those to this tool (if it was to be used) for each of the targets. I think this also means `std/go` would need to be recreated from scratch each time (via `rm -rf` or equivalent), so nothing left over from a previous build, perhaps for a different architecture (or version of Go), would get picked up. Possibly the tool itself should do this (when `--dir` is specified, which will be the typical use case).

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
* Document the code better
* Probably should remove all the `--dump` and maybe `--list` stuff if/when the tool is operating smoothly

## Sample Usage

This uses a "small" copy of the `src/net/` subdirectory in the Go source tree -- enough to quickly iterate over getting `LookupMX()` to look more like we might want it to. A full copy of that subdirectory is in `tests/big`. Note that this tool currently manages to work on the entire `golang/go/src` tree, though it sees multiple definitions of the same function (and complains about them -- they shouldn't be output, of course).

```
$ ./gostd2joker --dir $PWD/tests/small 2>&1 | grep -E -C5 '(lookupMX|queryEscape)'

FUNC net.LookupMX has: 
(defn ^[[{:host ^String Host, :pref ^Int Pref}] Error] LookupMX
  "LookupMX returns the DNS MX records for the given domain name sorted by preference."
  {:added "1.0"
   :go "lookupMX(name)"}
  [^String name])

FUNC net.LookupPort has: 
(defn ^[port err] LookupPort
  "LookupPort looks up the port for the given network and service."
--
FUNC url.QueryEscape has: 
(defn ^String QueryEscape
  "QueryEscape escapes the string so it can be safely placed
inside a URL query."
  {:added "1.0"
   :go "queryEscape(s)"}
  [^String s])

FUNC url.PathEscape has: 
(defn ^String PathEscape
  "PathEscape escapes the string so it can be safely placed
$
```

Note that `^[[{:host ..., :pref ...}] Error]` construct -- it indicates that `LookupMX()` returns a vector whose first element is itself a vector of maps with the indicated keys, and whose second element is of type `Error`.

I'm not sure that syntax really works, though -- because it doesn't seem to distinguish between a one-element vector and a vector of multiple elements of the same type (of the one element listed).
