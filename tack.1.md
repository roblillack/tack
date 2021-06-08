% TACK(1)
% Robert Lillack
% June 2021

# NAME

tack - a static site generator for the long run

# SYNOPSIS

**tack** [*ACTION*] [*SITEDIR*]

# DESCRIPTION

Tack reads website sources, like Mustache HTML templates, Markdown content markup, and YAML site & page variables, and “tacks them together” to create a static website that can easily be hosted anywhere.

The tool is completely self-contained and has no runtime dependencies. This ensures that updates to the websites you are creating now are still easily possible do a few years down the road.

# ACTIONS

**tack**
: Tack the site together into the folder `output`. This is the default action, if no verb is specified.

**serve**
: Tack the site together and start a web server on port `8080` which can be used to get a live preview of the tacked website. Changes to the source files (content, templates, assest, ...) are re-tacked and reflected in the served site automatically.

**help**
: Display a friendly help message.

# SITE DIRECTORY

The optional _SITEDIR_ arguments refers to a directory that contains the sources
of the website project that should be built by tack.

If this argument is not specified, tack uses the current working directory.

A valid _SITEDIR_ contains:

- `content` directory with at least a single metadata (`*.yaml`) or markup (`*.md`) file
- `templates` subdirectory with at least a single template file (`*.mustache`)
- Optionally: a `public` subdirectory with static files
- Optionally: a `site.yaml` metadata file to define some site variables

# EXIT STATUS

Tack returns a non-zero exit code if tacking the website was not successful due to being unable to read or process any of the input files or if the `output` directory cannot be written to.

# BUGS

To report bugs, please go to create a ticket at https://github.com/roblillack/tack/issues

# SEE ALSO

jekyll(1)
