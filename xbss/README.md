xbss
====

Purpose
-------
Capture screenshots from debug enabled first generation Xbox consoles.

Install
-------
```
go install github.com/dstien/dutils/xbss
```

Use
---
```
xbss [-f filename.png] [-v] host
```

A filename with the format `xbss-2006-01-02_15-04-05.000.png` is generated if the `-f` argument is not set.

The output filename is printed to stdout and can be used to view the result:
```
$ xbss 192.168.0.42 | xargs geeqie
```

License
-------
[CC0 - Public domain](http://creativecommons.org/publicdomain/zero/1.0/)

Contact
-------
daniel@stien.org
