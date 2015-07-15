package kinesumer

import (
	"github.com/stretchr/testify/mock"
)

type KinesumerAPIMock struct {
	mock.Mock
}

func (m *KinesumerAPIMock) Begin() error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *KinesumerAPIMock) End() {
	m.Called()
}
func (m *KinesumerAPIMock) Records() <-chan *KinesisRecord {
	ret := m.Called()

	var r0 <-chan *KinesisRecord
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(<-chan *KinesisRecord)
	}

	return r0
}
