xbreboot
========

Purpose
-------
Remotely reboot debug enabled first generation Xbox consoles.

Install
-------
```
go install github.com/dstien/dutils/xbreboot
```

Use
---
```
xbreboot [-warm] [-v] host
```

Use the `-warm` flag to return to the dashboard without reloading the BIOS. No output is printed on successful execution unless the `-v` verbosity flag is set.

License
-------
[CC0 - Public domain](http://creativecommons.org/publicdomain/zero/1.0/)

Contact
-------
daniel@stien.org
