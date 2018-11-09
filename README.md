# gostd2joker

Idea here is for people who build Joker to optionally run (a future version of) this tool against a Go source tree, which _must_ correspond to the version of Go they use to build Joker itself, to populate `joker/std/go/`. Further, the build parameters (`$GOOS`, `$GOARCH`, etc.) must match -- so `build-all.sh` would have to pass those to this tool (if it was to be used) for each of the targets. I think this also means `joker/std/go` would need to be recreated from scratch each time (via `rm -rf` or equivalent), so nothing left over from a previous build, perhaps for a different architecture (or version of Go), would get picked up. Possibly the tool itself should do this (when `--populate` is specified, which will be the typical use case).

At the moment, this is just a proof of concept, focusing on `net.LookupMX()`. E.g. run it like this:

```
$ ./gostd2joker --source ~/github/golang/go/src 2>&1 | less  # --source is now a required option
```

Then page through it. Code snippets intended for e.g. `joker/std/go/net.joke` are currently just printed to `stdout`, making iteration much easier. (Or, specify `--populate <dir>` to get all the individual `*.joke` and `*.go` files in `joker/std/go/`.)

Anything not supported results in either a `panic` or, more often, the string `ABEND` along with some kind of explanation. The latter is used to auto-detect a non-convertible function, in which case the snippet(s) are still output, but commented-out, so it's easy to see what's missing and (perhaps) why.

Among things to do to "productize" this:

* IN PROGRESS: Add `--populate <dir>` option to specify the `joker/std/go/` directory into which to (over-?)write actual code
* IN PROGRESS: Generate imports and such properly
* NEEDED?: Change `generate-std.joke` to support this tool's output (this avoids the tool having to generate `*_native.go` snippets or files in many, if not all, cases)
* Might have to replace the current ad-hoc tracking of Go packages with something that respects `import` and the like
* SOMEWHAT DONE: Have tool generate more-helpful docstrings than just copying the ones with the Go functions -- e.g. the return types, maybe decorated with extra information?
* Explain how to use the tool in Joker's `README.md`
* Document the code better
* Assess performance impact (especially startup time) on Joker, and mitigate as appropriate

## Sample Usage

This uses a "small" copy of the `golang/go/src/net/` subdirectory in the Go source tree -- enough to quickly iterate over getting `LookupMX()` to look more like we might want it to. A full copy of that subdirectory is in `tests/big`. Note that this tool currently manages to work on the entire `golang/go/src/` tree, though it sees multiple definitions of the same function (and complains about them -- they shouldn't be output, of course).

```
$ ./gostd2joker --source $PWD/tests/small 2>&1 | grep -i -E -C20 '(lookupMX|queryEscape)'

JOKER FUNC net.LookupMX has:
(defn LookupMX
  "LookupMX returns the DNS MX records for the given domain name sorted by preference.\nGo return type: ([]*MX, error)\nJoker return type: [(vector-of {:Host ^String, :Pref ^Int}) Error]"
  {:added "1.0"
   :go "lookupMX(name)"}
  [^String name])

...

JOKER FUNC url.QueryEscape has:
(defn ^String QueryEscape
  "QueryEscape escapes the string so it can be safely placed\ninside a URL query.\nGo return type: string\nJoker return type: String"
  {:added "1.0"
   :go "url.QueryEscape(s)"}
  [^String s])

...

GO FUNC net.LookupMX has:
func lookupMX(name string) Object {
	res1, res2 := net.LookupMX(name)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		map2 := EmptyArrayMap()
		map2.Add(MakeKeyword("Host"), MakeString((*elem1).Host))
		map2.Add(MakeKeyword("Pref"), MakeInt(int((*elem1).Pref)))
		vec1 = vec1.Conjoin(map2)
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}
```

Note that `^[(vector-of {:Host ^String, :Pref ^Int}) Error]` construct -- it indicates that `LookupMX()` returns a vector whose first element is itself a vector of maps with the indicated keys, and whose second element is of type `Error`.

## Run Tests

The `test.sh` script runs tests against a small, then larger, then
(optionally) full, copy of Go 1.11's `golang/go/src/` tree.

If you have a copy of the Go source tree available, define the `GOSRC` environment variable to point to its `src/` subdirectory, ideally via a the relative symlink `../GOSRC`, to be compatible with existing test output. E.g.:

```
$ ln -s ~/golang/go/src ../GOSRC  # If $GOSRC is undefined, ../GOSRC will be tried
```

(This same environment variable might someday be "respected" by `gostd2joker` and even `joker/run.sh` someday.)

Then, invoke `test.sh` either with no options, or with `--on-error :` to run the `:` (`true`) command when it detects an error (the default being `exit 99`).

E.g.:

```
$ ./test.sh
$
```

The script currently runs tests in this order:

1. `tests/small`
2. `tests/big`
3. (If `$GOSRC` is non-null and points to a directory) `$GOSRC`

After each test it runs, it uses `git diff` to compare the resulting `.gold` file with the checked-out version and, if there are any differences, it runs the command specified via `--on-error` (again, the default is `exit 99`, so the script will exit as soon as it sees a failing test).

NOTE: `$GOSRC` can now be pointed to a symlink, and `tests/GOSRC.out` has been rebuilt with `GOSRC=../GOSRC`, with that being a symlink (one directory level above `gostd2joker` itself) to the Go source tree. This allows me to easily run on different machines and OSes without having tons of needless differences due to absolute pathnames being different (some machines use `/home`, others `/Users`, to hold home directories).

ALSO NOTE: The Go standard library is customized per system architecture and OS, and `gostd2joker` picks up these differences via its use of Go's build-related packages. As a result, differences in `tests/GOSRC.out`, and other tests likely to come, are to be expected when run on a different arch/OS than the latest checked-in version.
