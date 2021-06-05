CHANGELOG

next:

- Fix crash when not called from within a site directory
- `tack` and `serve` verbs allow specifying a site directory to work with
- `serve` will detect removed or newly added files
- Add timestamps to logging output for `tack serve`
- Automatically disables all browser caching.
- Startup time is 10x faster. Simple `tack tack` for a simple website is down from ~600ms to ~50ms on my old Mac.
