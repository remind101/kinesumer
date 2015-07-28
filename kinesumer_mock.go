package kinesumer

import (
	"github.com/stretchr/testify/mock"
)

type KinesumerMock struct {
	mock.Mock
}

func (m *KinesumerMock) Begin() error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *KinesumerMock) End() {
	m.Called()
}
func (m *KinesumerMock) Records() <-chan *KinesisRecord {
	ret := m.Called()

	var r0 <-chan *KinesisRecord
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(<-chan *KinesisRecord)
	}

	return r0
}
