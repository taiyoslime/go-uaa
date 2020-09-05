[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://www.repostatus.org/badges/latest/wip.svg)](https://www.repostatus.org/#wip) ![GitHub Workflow Status](https://img.shields.io/github/workflow/status/taiyoslime/niller/Go) [![Go Report Card](https://goreportcard.com/badge/taiyoslime/niller)](https://goreportcard.com/report/taiyoslime/niller) 

# niller
niller (nil + killer) is a static analysis tool that warns dangerous statement involving nil.

## Installation

```
$ go get -u github.com/taiyoslime/niller/cmd/niller
```

## Usage

```
$ go vet -vettool=$(which niller)
```

## Example 
```go
package a

import "errors"

type Test struct { val int }

func (t *Test) test() int { return t.val }

func CreateTest(cond bool) (*Test, error) {
	if cond {
		return &Test{}, nil
	} else {
		return nil, errors.New("err")
	}
}

func f() interface{} {
	var a = &Test{}
	var b *Test
	c, _ := CreateTest(true)
	d, err := CreateTest(true)
	if err != nil {
		return err
	}
	var e *Test
	if e, err = CreateTest(true); err != nil {
		return err
	}
	var (
		f *Test
	)

	a.test()
	b.test() // warns "b may be nil"
	c.test() // warns "c may be nil"
	d.test()
	e.test()
	f.test() // warns "f may be nil"
	return nil
}
```