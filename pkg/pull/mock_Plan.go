// Code generated by mockery v1.0.0. DO NOT EDIT.

package pull

import mock "github.com/stretchr/testify/mock"

// MockPlan is an autogenerated mock type for the Plan type
type MockPlan struct {
	mock.Mock
}

// InitFilter provides a mock function with given fields:
func (_m *MockPlan) InitFilter() Filter {
	ret := _m.Called()

	var r0 Filter
	if rf, ok := ret.Get(0).(func() Filter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Filter)
		}
	}

	return r0
}

// Steps provides a mock function with given fields:
func (_m *MockPlan) Steps() StepList {
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
