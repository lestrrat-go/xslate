go-xslate
=========

Attempt to port Perl5's Text::Xslate to Go

[![Build Status](https://travis-ci.org/lestrrat/go-xslate.png?branch=master)](https://travis-ci.org/lestrrat/go-xslate)

[![GoDoc](https://godoc.org/github.com/lestrrat/go-xslate?status.png)](https://godoc.org/github.com/lestrrat/go-xslate)

Description
===========

This is an attempt to port [Text::Xslate](https://github.com/xslate/p5-Text-Xslate) from Perl5 to Go.

Xslate is an extremely powerful virtual machine based template engine.

Why Would I Choose xslate over text/template?
=============================================

I believe there are at least two reasons you would choose Xslate over the basic
text/template or html/template packages:

*Template flexibility*

IMHO, the default TTerse syntax is much more expressive and flexible.
With WRAPPERs and INCLUDEs, it is possible to write a very module set of 
templates. YMMV

*Dynamic/Automatic Reloading*

By default Xslate expects that your template live in the file system -- i.e.
outside of your go code. While text/template expects that you manage loading
of templates yourself. Xslate handles all this for you. It searches for
templates in the specified path, does the compilation, and handles caching,
both on memory and on file system.

Xslate is also designed to allow you to customize this behavior: It should be
easy to create a template loader that loads from databases and cache into
memcached and the like.

Current Status
=======

Currently:

* I'm aiming for port of most of TTerse syntax
* See [VM Progress](https://github.com/lestrrat/go-xslate/wiki/VM-Progress) for what the this xslate virtual machine can handle
* VM TODO: macros
* Parser is about 80% finished.
* Compiler is about 60% finished.
* Pluggable syntax isn't implemented at all.

For simple templates, you can already do:

```go
package main

import (
  "log"
  "github.com/lestrrat/go-xslate"
)

func main() {
  xt := xslate.New()
  template := `Hello World, [% name %]!`
  output, err := xt.RenderString(template, xslate.Vars { "name": "Bob" })
  if err != nil {
    log.Fatalf("Failed to render template: %s", err)
  }
  log.Printf(output)
}
```

See [Supported Syntax (TTerse)](https://github.com/lestrrat/go-xslate/wiki/Supported-Syntax-(TTerse)) for what's currently available

Debugging
=========

Currently the [error reporting is a bit weak](https://github.com/lestrrat/go-xslate/issues/4). What you can do when you debug or send me bug reports is to give me a stack trace, and also while you're at it, run your templates with XSLATE_DEBUG=1 environment variable. This will print out the AST and ByteCode structure that is being executed.

Caveats
=======

Functions
---------

In Go, functions that are not part of current package namespace must be
qualified with a package name, e.g.:

    time.Now()

This works fine because you can specify this at compile time, but you can't
resolve this at runtime... which is a problem for templates. The way to solve
this is to register these functions as variables:

    template = `
      [% now() %]
    `
    tx.RenderString(template, xslate.Vars { "now": time.Now })

But this forces you to register these functions every time, as well as
having to take the extra care to make names globally unique.

    tx := xslate.New(
      functions: map[string]FuncDepot {
        // TODO: create pre-built "bundle" of these FuncDepot's
        "time": FuncDepot { "Now": time.Now }
      }
    )
    template := `
      [% time.Now() %]
    `
    tx.RenderString(template, ...)


Comparison Operators
--------------------

The original xslate, written for Perl5, has comparison operators for both
numeric and string ("eq" vs "==", "ne" vs "!=", etc). In go-xslate, there's
no distinction. Both are translated to the same opcode (XXX "we plan to", that is)

So these are the same:

    [% IF x == 1 %]...[% END %]
    [% IF x eq 1 %]...[% END %]


Accessing Fields
----------------

Only public struc fields are accessible from templates. This is a limitation of the Go language itself.
However, in order to allow smooth(er) migration from p5-Text-Xslate to go-xslate, go-xslate automatically changes the field name's first character to uppercase.

So given a struct like this:

```go
  x struct { Value int }
```

You can access `Value` via `value`, which is common in p5-Text-Xslate

```
  [% x.value # same as x.Value %]
```
