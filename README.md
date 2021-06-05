# tack [![Build Status](https://secure.travis-ci.org/roblillack/tack.png?branch=master)](http://travis-ci.org/roblillack/tack) [![GoDoc](http://godoc.org/github.com/roblillack/tack?status.png)](http://godoc.org/github.com/roblillack/tack)

**tack** is a static site generator for the long run.

The project's goal is to create and maintain a sustainable tool that does the
(arguably pretty easy) job of filling HTML templates with content well enough
now and in ten years when you come back to update that minimal website you own.

#### Installation

```
go get github.com/roblillack/tack
```

#### Usage

Create directory for your site somewhere using a structure like this:

```
mysite                     Your website project dir
├── content                Contains a subdir per page
│   ├── about-me           Page will be available at /about-me
│   │   ├── default.yaml   Page variables, page will use “default” template
│   │   ├── body.md        One page variable “content” will hold this files'
│   │   │                  content processed as HTML.
│   │   └── me.jpg         All files not recognized as metadata or markup will
|   |                      be regarded as assets and be copied to output as is.
│   ├── bikes              Another page, /bikes
│   │   └── body.md        Works, even if no other page variables are defined.
│   └── work               Again, another page: /work
│       └── serious.yaml   Different template used here.
├── templates
│   ├── default.mustache   The default template, used by /about-me and /bikes.
│   └── serious.mustache   Another template, used by /work
└── public                 Files in here will not be touched and will be copied
    ├── style.css          over to output/ as is.
    ├── logo.png
    └── js
        ├── main.js
        ├── tracker.js
        └── library.js
```

To create the static site in `output/`, just run

```
tack
```

from inside your site directory. Alternatively run:

```
tack serve
```

and open your browser at http://localhost:8080/ while working on the site.

Once you're done, copy over the content of `output/` to a hosting service of your choice.

#### Why?

#### Features that will not be part of future tack versions

There are lots of features that are more or less a standard part of static site
generators nowadays but don't really align well with the goals of the project and
therefor will not be added to tack.

- Plugin support
- Image resizing
- JavaScript transpilation
- JavaScript bundling
- JavaScript minification

#### Features that might be implemented as part of future tack versions

- Sitemap creation
- CSS transpilation (we used to have less support)
- TOML file metadata support
- Liquid template support
- More configuration options

#### License

[MIT/X11](https://github.com/roblillack/tack/blob/master/LICENSE).
