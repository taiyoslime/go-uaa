package a

import "errors"

type Test struct {
}

func (t *Test) test() int {
	return 1
}

func CreateTest(flag bool) *Test {
	if flag {
		return &Test{}
	} else {
		return nil
	}
}

func CreateTestWithErr(flag bool) (*Test, error) {
	if flag {
		return &Test{}, nil
	} else {
		return nil, errors.New("err")
	}
}

func hogefuga() (a, b int, c int){
	return 1, 1, 1
}


func f() interface{} {
	var a = &Test{}
	var b *Test
	var t = CreateTest(true)
	if t == nil {
		panic("")
	}
	var s = CreateTest(true)
	var aa, ab, ac = hogefuga()
	var (
		e *Test
		f *Test
	)
	g := CreateTest(true)
	h, err := CreateTestWithErr(true)
	if err != nil {
		return err
	}

	x := a.test()
	y := b.test() // want "b may be nil"
	xx := t.test()
	xy := e.test() // want "e may be nil"
	xz := f.test() // want "f may be nil"
	yx := s.test() // want "s may be nil"
	ga := g.test() // want "g may be nil"
	gb := h.test()
	return x + y + xx + aa + ab + ac + xy + xz + yx + ga + gb
}