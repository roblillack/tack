# CHANGELOG

## next

## v1.0.0 - 2021-06-06

- BREAKING: Support for CSS filters (ie. LESS support) was dropped!
- Added support for specifiying a site directory for the `tack` and `serve` commands.
- `tack serve` will now detect removed or newly added files.
- Added timestamps to logging output of `tack serve`
- Fixed crash when not called from within a site directory.
- Startup time is >10x faster. Calling `tack tack` for a simple website is down from ~600ms to ~50ms on my very old Mac.
- The software is rewritten in Go (compared to C#) for the following reasons:
  - Support for more platforms (FreeBSD, NetBSD, OpenBSD, and Windows in addition to Linux and macOS)
  - Maintenance & setting up build environment is _way_ easier with Go compared to .NET (even if this improved in the past 10 years)
  - Dependency tracking is easier and more reliable in Go
  - Runtime requirements are lower (memory and disk space footprint) for Go version
  - Go version is faster (esp. startup time)

## v0.5.1 - 2019-07-31

- Rewrites dependency handling for to use Nuget where applicable
- Rewrites build infrastructure to work with .NET Core
- First time binaries are provided as part of the release

## v0.5.0 - 2012-08-25

- First usable version
- Runs (using Mono) on MacOS X and Linux
- Includes the following features:
  - Tacking websites
  - Embedded development server
  - Mustache template support
  - YAML metadata support
  - LESS CSS transpilation
