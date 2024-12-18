# Internals

## Version

This document describes component versions:

* `urdt-ussd` `v0.5.0-beta`
* `go-vise` `v0.2.2`


## User profile data

All user profile items are stored under keys matching the user's session id, prefixed with the 8-bit value `git.defalsify.org/vise.git/db.DATATYPE_USERDATA` (32), and followed with a 16-big big-endian value subprefix.

For example, given the sessionId `+254123` and the key `git.grassecon.net/urdt-ussd/common.DATA_PUBLIC_KEY` (2) will be stored under the key:

```
0x322b3235343132330002

prefix   sessionid       subprefix
32       2b323534313233  0002
```

### Sub-prefixes

All sub-prefixes are defined as constants in the `git.grassecon.net/urdt-ussd/common` package. The constant names have the prefix `DATA_`

Please refer to inline godoc documentation for the `git.grassecon.net/urdt-ussd/common` package for details on each data item.
