package util

import "math/rand"

type Sampler struct {
	Options []uint
}

func (s *Sampler) Add(idx, count uint) {
	var i uint
	for i = 0; i < count; i++ {
		s.Options = append(s.Options, idx)
	}
}

func (s *Sampler) Sample() uint {
	return s.Options[rand.Intn(len(s.Options))]
}
