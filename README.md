# gogtags

[![Build Status](https://travis-ci.org/juntaki/gogtags.svg?branch=master)](https://travis-ci.org/juntaki/gogtags)
[![Coverage Status](https://coveralls.io/repos/github/juntaki/gogtags/badge.svg?branch=master)](https://coveralls.io/github/juntaki/gogtags?branch=master)

GNU global compatible source code tagging for golang

## How to use

~~~
juntaki@ubuntu ~/.g/s/g/j/gogtags> gogtags
juntaki@ubuntu ~/.g/s/g/j/gogtags> ls G*
GPATH  GRTAGS  GTAGS
juntaki@ubuntu ~/.g/s/g/j/gogtags> global -x main
main              254 main.go          func main() {
juntaki@ubuntu ~/.g/s/g/j/gogtags> global -rx dump
dump              133 main.go          func (g *global) dump() {
dump              168 main.go           g.dump()
dump              237 main.go                   g.dump()
juntaki@ubuntu ~/.g/s/g/j/gogtags> global -sx lineNumbers
lineNumbers        33 main.go           lineNumbers []int
lineNumbers        38 main.go           output := fmt.Sprintf("%d", c.lineNumbers[0]) // [0] must be exist
lineNumbers        39 main.go           for l := 1; l < len(c.lineNumbers); l++ {
lineNumbers        40 main.go                   diff := c.lineNumbers[l] - c.lineNumbers[l-1]
lineNumbers        40 main.go                   diff := c.lineNumbers[l] - c.lineNumbers[l-1]
lineNumbers       215 main.go                   r.lineNumbers = append(r.lineNumbers, pos.Line)
lineNumbers       215 main.go                   r.lineNumbers = append(r.lineNumbers, pos.Line)
lineNumbers       219 main.go                           lineNumbers: []int{pos.Line},
~~~
