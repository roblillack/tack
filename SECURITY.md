# Security Policy

## What is regarded a security defect?

With tack being a static site generator, naturally the attack surface is very
low: A static site generator is usually not run in any kind of production
environment.

Still, we want to ensure that users of the tool can trust it to not break their
CI or development systems and we therefore regard the following types of defects
a security issue:

- Writing to the filesystem outside of `SITE/output`
- Serving data which does not belong to the generated website when running `tack serve`

## Supported Versions

We'll only support the latest major version of tack with security updates. Currently this means:

| Version | Supported          |
| ------- | ------------------ |
| 1.2.x   | :white_check_mark: |
| 1.1.x   | :x:                |
| 1.0.x   | :x:                |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Feel free to report security defects using [our bug tracker](https://github.com/roblillack/tack/issues). If you'd rather report a security issue privately, you can do so by sending email to [@roblillack](https://github.com/roblillack): To get to my email address, just add the at sign between my given and family name and finish it of by adding .net!
