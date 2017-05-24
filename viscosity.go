package main

// V is the instance for calc viscosity
type V struct {
	base int
}

// calcMV returns `viscosity` for given n
func (v *V) calcMV(n int) (mv int) {
	// fmt.Printf("start calcMV for %d\n", n)
	mv = 1
	for {
		n = v.step(n)
		if n < 10 {
			break
		}
		mv++
	}
	return mv
}

// step returns `next step` integer for `viscosity` calculation.
// one step proceed following:
//   - For given n, multiply all digit
func (v *V) step(n int) (result int) {
	// fmt.Printf("start step for %d\n", n)
	var r int
	result = 1
	for {
		n, r = n/v.base, n%v.base
		result *= r
		if n == 0 {
			break
		}
	}
	return
}
