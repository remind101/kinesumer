package kinesumer

import (
	k "github.com/remind101/kinesumer/interface"
	"github.com/stretchr/testify/mock"
)

type CheckpointerMock struct {
	mock.Mock
}

func (m *CheckpointerMock) DoneC() chan<- *k.KinesisRecord {
	ret := m.Called()

	var r0 chan *k.KinesisRecord
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(chan *k.KinesisRecord)
	}

	return r0
}
func (m *CheckpointerMock) Begin(_a0 chan<- *k.KinesisRecord) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
func (m *CheckpointerMock) End() {
	m.Called()
}
func (m *CheckpointerMock) GetStartSequence(shardID *string) *string {
	ret := m.Called(shardID)

	var r0 *string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*string)
	}

	return r0
}
func (m *CheckpointerMock) Sync() {
	m.Called()
}
func (m *CheckpointerMock) TryAcquire(shardID *string) error {
	ret := m.Called(shardID)

	r0 := ret.Error(0)

	return r0
}
func (m *CheckpointerMock) Release(shardID *string) error {
	ret := m.Called(shardID)

	r0 := ret.Error(0)

	return r0
}
