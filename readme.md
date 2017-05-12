csv-clean
=========

`csv-clean` is a short command line tool to take a malformed csv file, and convert it to
[RFC 4180](https://www.ietf.org/rfc/rfc4180.txt) format.

It assumes the input is UTF-8 and lines are terminated with `\n` characters. 
Behaviour on non-UTF-8 input is undefined.

A general description of the RFC 4180 format is:

* fields may be quoted in `"` characters
* non-quoted fields may not include newlines (`\n`), commas (`,`) or double quote (`"`) characters
* quoted fields may include newlines and commas. Any double quotes within a field must be doubled (`""`)

## Corrections

This tool will correct the following errors in a range of conditions (including fields with newlines):

`unquoted field, with a " inside` -> `unquoted field," with a "" inside"`
`unquoted field,with a badly escpaed \" inside` -> `unquoted field,"with a badly escaped "" inside"`
`quoted field," with an unescaped " inside"` -> `quoted field," with an unescaped "" inside"`
`quoted field,"with a badly escpaed \" inside"` -> `quoted field,"with a badly escaped "" inside"`

It is unable to correct unquoted fields that erroneously start with a `"`.

