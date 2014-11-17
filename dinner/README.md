dinner
======

Purpose
-------
Ticker for Dovre Forvaltning funds.

Install
-------
```
go install github.com/dstien/dutils/dinner
```

Use
---
```
dinner [-f FUND] [-t TYPE] [-c CurrentFile] [-v]
```

Last processed date is stored in `CurrentFile` if set. Used to prevent unnecessary downloads for already fetched days. Run `dinner -h` to see available funds and formatting options.

A cron job can be used for polling:
```
*/10 9-12 * * 1-5 dinner -f DIN -t IRC -c ~/.dinner/din.current 1>> ~/.dinner/din.output
```

License
-------
[CC0 - Public domain](http://creativecommons.org/publicdomain/zero/1.0/)

Contact
-------
daniel@stien.org
