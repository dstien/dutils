vgknit
======

Purpose
-------
Produce knitting pattern from PNG files for VG's 2014 World Chess Championship knitting competition at http://magnusgenseren.vg.no/

Image dimensions must be 20x270 for pattern and 80x270 for main motif. Accepted colors are black, white and transparent (to make pattern visible behind main motif).

One pixel correspond to one stitch. Stitches are non-square (3:2), so scale your drawings accordingly.

Install
-------
```
go install github.com/dstien/dutils/vgknit
```

Use
---
```
vgknit pattern.png motif.png | xclip
```

Go to http://magnusgenseren.vg.no/ and paste result in your browser's JS console.

License
-------
[CC0 - Public domain](http://creativecommons.org/publicdomain/zero/1.0/)

Contact
-------
daniel@stien.org
