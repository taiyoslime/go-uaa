package a

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
	x := a.test()
	y := b.test() // want "b may be nil"
	xx := t.test()
	xy := e.test() // want "e may be nil"
	xz := f.test() // want "f may be nil"
	yx := s.test() // want "s may be nil"
	return x + y + xx + aa + ab + ac + xy + xz + yx
}