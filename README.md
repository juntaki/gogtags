# gogtags

[![Build Status](https://travis-ci.org/juntaki/gogtags.svg?branch=master)](https://travis-ci.org/juntaki/gogtags)

GNU GLOBAL compatible source code tagging for golang

## Installation

~~~
go get github.com/juntaki/gogtags
~~~

## GNU GLOBAL Installation for gogtags

GNU GLOBAL **with sqlite3** is required for reference.
https://www.gnu.org/software/global/

### Debian/Ubuntu
~~~
sudo apt install libncurses5-dev build-essential  # for ubuntu, build dependency
wget http://tamacom.com/global/global-6.5.7.tar.gz
tar xvf global-6.5.7.tar.gz
cd global-6.5.7
./configure --with-sqlite3
make
sudo make install
~~~

### Mac

~~~
brew install global -with-sqlite3
~~~

## How to use

~~~
gogtags -v
~~~

![screenshot1](https://github.com/juntaki/gogtags/blob/master/gogtags_screenshot1.gif?raw=true)


## Use with emacs helm-gtags or other editor plugin

Just use it as ususal, Generated tag is GNU GLOBAL compatible.

![screenshot1](https://github.com/juntaki/gogtags/blob/master/gogtags_screenshot2.gif?raw=true)
