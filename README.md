# encode.go
encode.go is a simple command line utility for encoding one or more
videos to be html5 friendly. The design is guided by
http://www.html5rocks.com/en/tutorials/video/basics and
http://diveintohtml5.info/video.html and empowered by [cli.go](https://github.com/codegangsta/cli).

## Overview
encode.go accepts a path and will recursively select, according to the value of --formats (default: mp4, webm,
ogg, ogv, wmv), and encode all matching files into Ogg Theora, MP4 and WebM formats. Encoded files are stored in the value of --destination (default: output). To generate a simple index.html file that contains the necessary video tags, include --html.

## Getting Started
To install encode.go, run:
```
$ go get github.com/timgremore/encode
```

For a list of options:
```
$ encode batch --help
```
