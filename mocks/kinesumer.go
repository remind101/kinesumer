package mocks

import (
	k "github.com/remind101/kinesumer/interface"
	"github.com/stretchr/testify/mock"
)

type Kinesumer struct {
	mock.Mock
}

func (m *Kinesumer) Begin() error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *Kinesumer) End() {
	m.Called()
}
func (m *Kinesumer) Records() <-chan *k.KinesisRecord {
	ret := m.Called()

	var r0 <-chan *k.KinesisRecord
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(<-chan *k.KinesisRecord)
	}

	return r0
}
