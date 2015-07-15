package kinesumer

import (
	"github.com/stretchr/testify/mock"
)

type ShardStateSyncMock struct {
	mock.Mock
}

func (m *ShardStateSyncMock) DoneC() chan *KinesisRecord {
	ret := m.Called()

	var r0 chan *KinesisRecord
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(chan *KinesisRecord)
	}

	return r0
}
func (m *ShardStateSyncMock) Begin() error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *ShardStateSyncMock) End() {
	m.Called()
}
func (m *ShardStateSyncMock) GetStartSequence(shardID *string) *string {
	ret := m.Called(shardID)

	var r0 *string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*string)
	}

	return r0
}
func (m *ShardStateSyncMock) Sync() {
	m.Called()
}
