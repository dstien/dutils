xbcp
====

Purpose
-------
Copy a local file to a debug enabled first generation Xbox console.

Install
-------
```
go install github.com/dstien/dutils/xbcp
```

Use
---
```
xbcp [-v] [sourcefile] [host:destfile]
```

Destination filename is on the format `host:X:\path\to\file`, where `host` is the IP or hostname of the Xbox console and X is the Xbox partition letter. If the last character is `/`, the local filename is used. The destination directory must exist.

Example:
```
$ xbxp ~/myfile 192.168.0.42:'Z:\hisfile'
```

TODO
----
* Check if remote directory exists by using `getfileattributes` command for better error handling.
* Copy from Xbox to local machine.

License
-------
[CC0 - Public domain](http://creativecommons.org/publicdomain/zero/1.0/)

Contact
-------
daniel@stien.org
