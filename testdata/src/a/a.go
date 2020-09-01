package a

type Test struct {
}

func (t *Test) test() int {
	return 1
}

func f() interface{} {
	var a = &Test{}
	var b *Test
	x := a.test()
	y := b.test() // want "b may be nil"
	return x + y
}