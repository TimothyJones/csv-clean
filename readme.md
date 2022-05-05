csv-clean
=========

`csv-clean` is a short command line tool to take a malformed csv file, and convert it to
[RFC 4180](https://www.ietf.org/rfc/rfc4180.txt) format.

[![Build Status](https://travis-ci.org/TimothyJones/csv-clean.svg?branch=master)](https://travis-ci.org/TimothyJones/csv-clean)

It assumes the input is UTF-8 and lines are terminated with `\n` characters. 
Behaviour on non-UTF-8 input is undefined.

It reads from standard input and writes to standard output

Usage: 

    cat <malformed.csv> | csv-clean > <clean.csv>
    cat <malformed.csv> | csv-clean -fixSlashedQuotes=false > <clean.csv>

Installation:

    go get github.com/TimothyJones/csv-clean


A general description of the RFC 4180 format is:

* fields may be quoted in `"` characters
* non-quoted fields may not include newlines (`\n`), commas (`,`) or double quote (`"`) characters
* quoted fields may include newlines and commas. Any double quotes within a field must be doubled (`""`)

## Corrections

This tool will correct the following errors in a range of conditions (including fields with newlines):

    from: unquoted field, with a " inside
      to: unquoted field," with a "" inside"

    from: unquoted field,with a badly escpaed \" inside
      to: unquoted field,"with a badly escaped "" inside"

    from: quoted field," with an unescaped " inside"
      to: quoted field," with an unescaped "" inside"

    from: quoted field,"with a badly escpaed \" inside"
      to: quoted field,"with a badly escaped "" inside"

It is unable to correct unquoted fields that erroneously start with a `"`.

You can disable the correction of `\"` to `""` with `-fixSlashedQuotes=false`

## See also

For an easy way to check that the input matches RFC-4180, see [csv-check](https://github.com/TimothyJones/csv-check)
