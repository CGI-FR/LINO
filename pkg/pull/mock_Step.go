// Code generated by mockery v1.0.0. DO NOT EDIT.

package pull

import mock "github.com/stretchr/testify/mock"

// MockStep is an autogenerated mock type for the Step type
type MockStep struct {
	mock.Mock
}

// Cycles provides a mock function with given fields:
func (_m *MockStep) Cycles() CycleList {
	ret := _m.Called()

	var r0 CycleList
	if rf, ok := ret.Get(0).(func() CycleList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(CycleList)
		}
	}

	return r0
}

// Entry provides a mock function with given fields:
func (_m *MockStep) Entry() Table {
	ret := _m.Called()

	var r0 Table
	if rf, ok := ret.Get(0).(func() Table); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Table)
		}
	}

	return r0
}

// Follow provides a mock function with given fields:
func (_m *MockStep) Follow() Relation {
	ret := _m.Called()

	var r0 Relation
	if rf, ok := ret.Get(0).(func() Relation); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Relation)
		}
	}

	return r0
}

// Index provides a mock function with given fields:
func (_m *MockStep) Index() uint {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}

// NextSteps provides a mock function with given fields:
func (_m *MockStep) NextSteps() StepList {
	ret := _m.Called()

	var r0 StepList
	if rf, ok := ret.Get(0).(func() StepList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(StepList)
		}
	}

	return r0
}

// Relations provides a mock function with given fields:
func (_m *MockStep) Relations() RelationList {
	ret := _m.Called()

	var r0 RelationList
	if rf, ok := ret.Get(0).(func() RelationList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(RelationList)
		}
	}

	return r0
}