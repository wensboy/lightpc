# Q&A

> Project · Q&A

Questions that came up while building the current version, with the answer that settled them.

## Why does a missing engine dependency stop everything?

`EngineRuntime.Run` resolves the dependency graph before any `Init`. An engine that depends on a sink the
configuration may leave empty (`log.direct`) makes the whole graph unresolvable, so every engine after it is
skipped. Depend only on what always exists, and use `slog.Default()` for diagnostics.

## Why is the settings file nested while the engine addresses dotted keys?

Leaves are addressed as `appearence.mode`, which keeps a write independent of its siblings. The file must stay
nested to match the schema, so `flatten` and `nest` convert at the boundary. Changing one side requires
changing the other.

## Why did the documentation panel show nothing with a clean log?

`path.Join` cleans its result, so `embed://data/docs/beta` plus `meta.json` became `embed:/…`, which the path
parser rejects. The error was swallowed by a `continue`. Use `internal/path.Join` for resource paths.

## Why does a wide table or a long code line push the dialog out?

A flex child defaults to `min-width: auto`, meaning "no smaller than my content". Declaring `min-w-0` along the
whole chain is what lets `max-w` act as a ceiling. The same applies to grid children.

## Why do markdown links not open inside the app?

External links are sanitised with `target="_blank"` on purpose: the reader must not navigate away from the
dialog. Internal document links stay plain text because there is no router entry for them yet.
