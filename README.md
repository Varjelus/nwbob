nwbob
---

nwbob is a go-gettable **build tool** for **NW.js** written in Golang. It's modeled after [nodebob](https://github.com/geo8bit/nodebob) ([my fork](https://github.com/Varjelus/nodebob))


Tech
-
nwbob uses a number of open source projects to work properly:

* [Anolis Resourcer](http://anolis.codeplex.com/) - a windows resource editor, v0.9.0
* [NW.js](http://nwjs.io/) - v0.12.3-win64


Installation
-
* [Install Go](https://golang.org/doc/install) (version >=1.5)
* Run `go get github.com/Varjelus/nwbob`
* Run `go install github.com/Varjelus/nwbob`

Usage
-
* Make sure `$GOPATH/bin` is in your path
* `cd` to your NW.js project directory containing `package.json`
* Run `nwbob`

For help about `nwbob` command arguments, run `nwbob -help`

Version
-
0.1.0. Tested on `Windows 10 64-bit`.

License
-
MIT
