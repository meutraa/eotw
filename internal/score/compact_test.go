package score

import (
	"testing"
	"time"

	"git.lost.host/meutraa/eotw/internal/game"
)

var compactTests = map[*([]game.Input)]([]InputsCompact){
	{}: {},
	{{Index: 0, HitTime: 100}, {Index: 3, HitTime: 200}}: {
		{Index: 0, Times: []time.Duration{100}},
		{Index: 1, Times: []time.Duration{}},
		{Index: 2, Times: []time.Duration{}},
		{Index: 3, Times: []time.Duration{200}},
	},
	{{Index: 1, HitTime: 2}, {Index: 1, HitTime: 1}}: {
		{Index: 0, Times: []time.Duration{}},
		{Index: 1, Times: []time.Duration{2, 1}},
	},
}

func TestCompactInputs(t *testing.T) {
	equal := func(p, q []InputsCompact) bool {
		if len(p) != len(q) {
			return false
		}
		for i := 0; i < len(p); i++ {
			pi, qi := p[i], q[i]
			if pi.Index != qi.Index {
				return false
			}
			if len(pi.Times) != len(qi.Times) {
				return false
			}
			for j := 0; j < len(pi.Times); j++ {
				if pi.Times[j] != qi.Times[j] {
					return false
				}
			}
		}
		return true
	}

	for in, expected := range compactTests {
		out := compactInputs(in)
		if !equal(out, expected) {
			t.Log("out     ", out)
			t.Log("expected", expected)
			t.Fail()
		}
	}
}

func TestUncompactInputs(t *testing.T) {
	equal := func(pp, qp *[]game.Input) bool {
		if nil == pp && nil == qp {
			return true
		} else if nil == pp || nil == qp {
			return false
		}

		p, q := *pp, *qp
		if len(p) != len(q) {
			return false
		}
		for i := 0; i < len(p); i++ {
			if p[i].Index != q[i].Index {
				return false
			}
			if p[i].HitTime != q[i].HitTime {
				return false
			}
		}
		return true
	}

	for expected, in := range compactTests {
		out := uncompactInputs(in)
		if !equal(out, expected) {
			t.Log("in      ", in)
			t.Log("expected", expected)
			t.Fail()
		}
	}
}
