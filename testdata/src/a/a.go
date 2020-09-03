package a

import "errors"

type Test struct {
}

func (t *Test) test() int {
	return 1
}

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
	/*
	var g *Test
	g = &Test{}
	g.test()
	*/

	a.test()
	b.test() // want "b may be nil"
	c.test() // want "c may be nil"
	d.test()
	e.test()
	f.test() // want "f may be nil"

	return nil
}