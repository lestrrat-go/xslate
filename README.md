go-xslate
=========

Attempt to port Perl5's Text::Xslate to Go

Description
===========

This is an attempt to port [Text::Xslate](https://github.com/xslate/p5-Text-Xslate) from Perl5 to Go.
HOWEVER, although the author has 10+ yrs experience programming, he has absolutely no experience developing virtual machines, compilers, et al. Your help is much, much, much needed (note: "appreciated" is an understatement. it's "needed")

Roadmap
=======

Currently I'm aiming for port of most of TTerse syntax

Caveats
=======

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
