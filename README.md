go-xslate
=========

Attempt to port Perl5's Text::Xslate to Go

[![Build Status](https://travis-ci.org/lestrrat/go-xslate.png?branch=master)](https://travis-ci.org/lestrrat/go-xslate)

Description
===========

This is an attempt to port [Text::Xslate](https://github.com/xslate/p5-Text-Xslate) from Perl5 to Go.
HOWEVER, although the author has 10+ yrs experience programming, he has absolutely no experience developing virtual machines, compilers, et al. Your help is much, much, much needed (note: "appreciated" is an understatement. it's "needed")

Current Status
=======

Currently:

* I'm aiming for port of most of TTerse syntax
* I'm working on the Virtual Machine portion
* VM currently supports: print\_raw, variable subtitution, arithmetic (add, subtract, multiply, divide), if/else conditionals, "for x in list" loop, simple method calls
* VM TODO: loops, macros, stuff involving external templates
* Parser is currently not finished.


Caveats
=======

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
