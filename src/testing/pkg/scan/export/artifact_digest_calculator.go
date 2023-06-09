// Code generated by mockery v2.22.1. DO NOT EDIT.

package export

import (
	digest "github.com/opencontainers/go-digest"

	mock "github.com/stretchr/testify/mock"
)

// ArtifactDigestCalculator is an autogenerated mock type for the ArtifactDigestCalculator type
type ArtifactDigestCalculator struct {
	mock.Mock
}

// Calculate provides a mock function with given fields: fileName
func (_m *ArtifactDigestCalculator) Calculate(fileName string) (digest.Digest, error) {
	ret := _m.Called(fileName)

	var r0 digest.Digest
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (digest.Digest, error)); ok {
		return rf(fileName)
	}
	if rf, ok := ret.Get(0).(func(string) digest.Digest); ok {
		r0 = rf(fileName)
	} else {
		r0 = ret.Get(0).(digest.Digest)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(fileName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewArtifactDigestCalculator interface {
	mock.TestingT
	Cleanup(func())
}

// NewArtifactDigestCalculator creates a new instance of ArtifactDigestCalculator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewArtifactDigestCalculator(t mockConstructorTestingTNewArtifactDigestCalculator) *ArtifactDigestCalculator {
	mock := &ArtifactDigestCalculator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
