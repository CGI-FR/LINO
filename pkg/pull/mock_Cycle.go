// Code generated by mockery v1.0.0. DO NOT EDIT.

package pull

import mock "github.com/stretchr/testify/mock"

// MockCycle is an autogenerated mock type for the Cycle type
type MockCycle struct {
	mock.Mock
}

// Len provides a mock function with given fields:
func (_m *MockCycle) Len() uint {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}

// Relation provides a mock function with given fields: idx
func (_m *MockCycle) Relation(idx uint) Relation {
	ret := _m.Called(idx)

	var r0 Relation
	if rf, ok := ret.Get(0).(func(uint) Relation); ok {
		r0 = rf(idx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Relation)
		}
	}

	return r0
}
