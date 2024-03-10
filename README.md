# gogtags

[![test](https://github.com/juntaki/gogtags/actions/workflows/go.yml/badge.svg)](https://github.com/juntaki/gogtags/actions/workflows/go.yml)

GNU GLOBAL compatible source code tagging for golang

## Installation

~~~
go install github.com/juntaki/gogtags@latest
~~~

## GNU GLOBAL Installation for gogtags

GNU GLOBAL **with sqlite3** is required for reference.
https://www.gnu.org/software/global/

### Mac

~~~
brew install global
~~~

## How to use

~~~
gogtags -v
~~~

![screenshot1](https://github.com/juntaki/gogtags/blob/master/gogtags_screenshot1.gif?raw=true)


## Use with emacs helm-gtags or other editor plugin

Just use it as usual, Generated tag is GNU GLOBAL compatible.

![screenshot1](https://github.com/juntaki/gogtags/blob/master/gogtags_screenshot2.gif?raw=true)
