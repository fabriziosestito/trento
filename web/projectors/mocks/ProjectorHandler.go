// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gorm "gorm.io/gorm"
)

// ProjectorHandler is an autogenerated mock type for the ProjectorHandler type
type ProjectorHandler struct {
	mock.Mock
}

// GetName provides a mock function with given fields:
func (_m *ProjectorHandler) GetName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Project provides a mock function with given fields: db, data
func (_m *ProjectorHandler) Project(db *gorm.DB, data interface{}) error {
	ret := _m.Called(db, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(*gorm.DB, interface{}) error); ok {
		r0 = rf(db, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Query provides a mock function with given fields: lastSeenIndex
func (_m *ProjectorHandler) Query(lastSeenIndex uint64) (interface{}, uint64, error) {
	ret := _m.Called(lastSeenIndex)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(uint64) interface{}); ok {
		r0 = rf(lastSeenIndex)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func(uint64) uint64); ok {
		r1 = rf(lastSeenIndex)
	} else {
		r1 = ret.Get(1).(uint64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(uint64) error); ok {
		r2 = rf(lastSeenIndex)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
