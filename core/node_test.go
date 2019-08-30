package core

import "github.com/stretchr/testify/mock"

type NodeMock struct {
	mock.Mock
}

func (m *NodeMock) IncreasePassRequest(count uint64) {
	m.Called(count)
	return
}
func (m *NodeMock) IncreaseBlockRequest(count uint64) {
	m.Called(count)
	return
}
func (m *NodeMock) IncreaseErrorRequest(count uint64) {
	m.Called(count)
	return
}
func (m *NodeMock) IncreaseSuccessRequest(count uint64) {
	m.Called(count)
	return
}
func (m *NodeMock) TotalRequest() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) TotalPass() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) TotalBlock() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) TotalSuccess() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) TotalError() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) RequestQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) PassQps() uint64 {
	args := m.Called()
	return uint64(*args.Get(0).(*int))
}

func (m *NodeMock) BlockQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) SuccessQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) ErrorQps() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) AddRt(rt uint64) {
	m.Called(rt)
	return
}
func (m *NodeMock) AvgRt() uint64 {
	args := m.Called()
	return uint64(args.Int(0))
}
func (m *NodeMock) IncreaseGoroutineNum() {
	m.Called()
	return
}
func (m *NodeMock) DecreaseGoroutineNum() {
	m.Called()
	return
}

func (m *NodeMock) CurGoroutineNum() int64 {
	args := m.Called()
	return int64(*args.Get(0).(*int))
}
