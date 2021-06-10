# Minimal website with blog

This is a minimal website which showcases the following features:

- Using content pages
- Using templates
- Using page variables
- Using static assets (CSS)

- Index page is part of the navigation

  This means, that the page available at `/` is not the _parent_ of the pages at `/products` and `/about` but at the same level.

- Renaming pages

  By default the `name` of a page is derived by title-casing the `slug` which is taken from the folder name). You can override this name, by
  setting a `name` page variable, which is done in `default.yaml` here, to
  call the index page `Home`.
