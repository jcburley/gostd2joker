# gostd2joker

Idea here is for people who build Joker to optionally run (a future version of) this tool against a Go source tree, which _must_ correspond to the version of Go they use to build Joker itself, to populate `joker/std/go/`. Further, the build parameters (`$GOOS`, `$GOARCH`, etc.) must match -- so `build-all.sh` would have to pass those to this tool (if it was to be used) for each of the targets. I think this also means `joker/std/go` would need to be recreated from scratch each time (via `rm -rf` or equivalent), so nothing left over from a previous build, perhaps for a different architecture (or version of Go), would get picked up. Possibly the tool itself should do this (when `--populate` is specified, which will be the typical use case).

At the moment, this is just a proof of concept, focusing on `net.LookupMX()`. E.g. run it like this:

```
$ ./gostd2joker --source ~/github/golang/go/src 2>&1 | less  # --source is now a required option
```

Then page through it. Code snippets intended for e.g. `joker/std/go/net.joke` are currently just printed to `stdout`, making iteration much easier.

Anything not supported results in either a `panic` or, more often, the string `ABEND` along with some kind of explanation. The latter is used to auto-detect a non-convertible function, in which case the snippet(s) are still output, but commented-out, so it's easy to see what's missing and (perhaps) why.

Among things to do to "productize" this:

* Add `--populate <dir>` option to specify the `joker/std/go/` directory into which to (over-?)write actual code
* Generate imports and such properly
* Change `generate-std.joke` to support this tool's output (this avoids the tool having to generate `*_native.go` snippets or files in many, if not all, cases)
* Might have to replace the current ad-hoc tracking of Go packages with something that respects `import` and the like
* Have tool generate more-helpful docstrings than just copying the ones with the Go functions -- e.g. the return types, maybe decorated with extra information?
* Explain how to use the tool in Joker's `README.md`
* Document the code better
* Probably should remove all the `--dump` and maybe `--list` stuff if/when the tool is operating smoothly
* Assess performance impact (especially startup time) and mitigate as appropriate

## Sample Usage

This uses a "small" copy of the `golang/go/src/net/` subdirectory in the Go source tree -- enough to quickly iterate over getting `LookupMX()` to look more like we might want it to. A full copy of that subdirectory is in `tests/big`. Note that this tool currently manages to work on the entire `golang/go/src/` tree, though it sees multiple definitions of the same function (and complains about them -- they shouldn't be output, of course).

```
$ ./gostd2joker --source $PWD/tests/small 2>&1 | grep -E -C5 '(lookupMX|queryEscape)'

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

I'm not sure that syntax really works, though -- because it doesn't seem to distinguish between a one-element vector and a vector of multiple elements of the same type (of the one element listed). I think Clojure itself would specify a Java class name instead; not sure to what that would translate in Joker.

## Run Tests

The `test.sh` script runs tests against a small, then larger, then
(optionally) full, copy of Go 1.11's `golang/go/src/` tree.

If you have a copy of the Go source tree available, define the `GOSRC` environment variable to point to its `src/` subdirectory. E.g.:

```
$ export GOSRC=~/golang/go/src
```

(This same environment variable might someday be "respected" by `gostd2joker` and even `joker/run.sh` someday.)

Then, invoke `test.sh` either with no options, or with `--on-error :` to run the `:` (`true`) command when it detects an error (the default being `exit 99`).

E.g.:

```
$ ./test.sh
$
```

The script currently runs tests in this order:

1. `tests/small/net`
2. `tests/big/net`
3. (If `$GOSRC` is non-null and points to a directory) `$GOSRC`

After each test it runs, it uses `git diff` to compare the resulting `.gold` file with the checked-out version and, if there are any differences, it runs the command specified via `--on-error` (again, the default is `exit 99`, so the script will exit as soon as it sees a failing test).
