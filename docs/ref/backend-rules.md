# Backend Rules

> Core purpose: Define the rules that backend code and design abstractions must follow.

## Must

1. Every struct field that can be exposed as an option must use a `withxxx` setter that returns the original reference. Example: `config.go:9.-29`.
2. The domain/ directory must contain only strongly domain-related declarations: const, struct, **interface**, and var. Related implementations belong in internal/, service/, engine/, or comparable directories.
3. Doc comments and file-top comments must be written entirely in English.
4. When implementing a requirement, prefer the standard library first, then a third-party library that is not newly released and has a substantial star count, and only as a last resort implement the algorithm yourself in an appropriate directory.
5. Package names that collide with the standard library must be disambiguated with a domain-qualified import alias written in lowerCamelCase.
6. Code annotated with a TODO comment must be implemented strictly according to the described steps, with attention to potential vulnerabilities. The implementation approach is unconstrained, but the function signature must not be changed.
7. File names within a package must be strongly semantic. Follow the less-is-more principle and avoid excessive fragmentation.

## Must Not

1. Add excessive explanatory comments that do not record a design decision.
2. Extract a function for logic that is confirmed to be called only once within a package and has no future reuse potential, such as a one-off constructor that merely creates a single struct instance.

## SHOULD

1. Prefer JSON for configuration files and YAML for structural files. Every such configuration file must have a corresponding schema file under `data/conf/schema/`: a configuration file named `example.{json|yaml}` maps to `example.schema.json`, which must conform to the JSON Schema standard defined by the RFC.
2. Files under data/template are typically Go `text/template` templates, used for dynamic configuration interpolation or for implementing complex requirements. Use `tpl` as the file extension.
3. Declare all `embed.FS` values in main.go wherever possible, so that embedded assets are not scattered across multiple locations and remain easy to manage centrally.
4. Under `internal/cast`, `cast.go` defines reusable cast functions for complex structures, and `fallback.go` defines reusable cast functions with callbacks (default values). ~~In general, conversions should be handled directly with `spf13/cast`, and any wrapper functions should also be built on top of it.~~
5. Use `.` to separate key parts when the key points to a valid value, and `::` when there is no corresponding value or the value is invalid. When a key has distinct semantic segments, use `.` within a segment and `:` between segments.
6. Keys used to look up a value do not need to be centralized in const declarations for the sake of avoiding hardcoding; keys that are not used for value lookup should be. For example, a config lookup key can be a literal string, while an event name should be centralized as a const.

## SHOULD NOT
