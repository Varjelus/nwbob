nodebob-go
---

nodebob-go is a go-gettable **build tool** for **NW.js** written in Golang. It's modeled after [nodebob](https://github.com/geo8bit/nodebob) ([my fork](https://github.com/Varjelus/nodebob))


Tech
-
nodebob-go uses a number of open source projects to work properly:

* ~~7-zip~~ [Replaced with my Go lib](https://github.com/Varjelus/archivist)
* Anolis Resourcer - a windows resource editor, v0.9.0
* NW.js - v1.2.0-win64

Version
-
0.1.0. Tested on `Windows 10 64-bit`.

Quick start
-
**Install**

`git clone https://github.com/Varjelus/nodebob-go.git nodebob-go`

`cd nodebob-go`

`go build`

**Run**

`cd` to a directory containing NW.js `package.json`

Execute `path/to/nodebob-go/nodebob-go.exe`

License
-
MIT

[node-webkit]: http://nwjs.io/
[Anolis Resourcer]: http://anolis.codeplex.com/
