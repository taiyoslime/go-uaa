package a

import "errors"

type Test struct {
	val int
}

func (t *Test) test() int {
	return t.val
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

	var g *Test
	g = &Test{}

	var h *Test
	h = &Test{}
	h = nil

	a.test()
	b.test() // want "b may be nil"
	c.test() // want "c may be nil"
	d.test()
	e.test()
	f.test() // want "f may be nil"
	g.test()
	h.test() // want "h may be nil"

	return nil
}