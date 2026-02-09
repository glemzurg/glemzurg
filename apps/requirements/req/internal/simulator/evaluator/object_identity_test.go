package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type ObjectIdentityTestSuite struct {
	suite.Suite
}

func TestObjectIdentitySuite(t *testing.T) {
	suite.Run(t, new(ObjectIdentityTestSuite))
}

func (s *ObjectIdentityTestSuite) TestNewIdentityRegistry() {
	registry := NewIdentityRegistry()
	s.NotNil(registry)
	s.Equal(0, registry.Count())
}

func (s *ObjectIdentityTestSuite) TestGetOrAssign_NewRecord() {
	registry := NewIdentityRegistry()
	record := object.NewRecord()
	record.Set("name", object.NewString("test"))

	id := registry.GetOrAssign(record)
	s.NotEqual(ObjectID(0), id)
	s.Equal(1, registry.Count())
}

func (s *ObjectIdentityTestSuite) TestGetOrAssign_SameRecordReturnsSameID() {
	registry := NewIdentityRegistry()
	record := object.NewRecord()
	record.Set("name", object.NewString("test"))

	id1 := registry.GetOrAssign(record)
	id2 := registry.GetOrAssign(record)

	s.Equal(id1, id2)
	s.Equal(1, registry.Count())
}

func (s *ObjectIdentityTestSuite) TestGetOrAssign_DifferentRecordsGetDifferentIDs() {
	registry := NewIdentityRegistry()

	record1 := object.NewRecord()
	record1.Set("name", object.NewString("first"))

	record2 := object.NewRecord()
	record2.Set("name", object.NewString("second"))

	id1 := registry.GetOrAssign(record1)
	id2 := registry.GetOrAssign(record2)

	s.NotEqual(id1, id2)
	s.Equal(2, registry.Count())
}

func (s *ObjectIdentityTestSuite) TestGetOrAssign_NilRecordReturnsZero() {
	registry := NewIdentityRegistry()
	id := registry.GetOrAssign(nil)
	s.Equal(ObjectID(0), id)
	s.Equal(0, registry.Count())
}

func (s *ObjectIdentityTestSuite) TestGetID_ExistingRecord() {
	registry := NewIdentityRegistry()
	record := object.NewRecord()
	record.Set("name", object.NewString("test"))

	assigned := registry.GetOrAssign(record)
	retrieved, ok := registry.GetID(record)

	s.True(ok)
	s.Equal(assigned, retrieved)
}

func (s *ObjectIdentityTestSuite) TestGetID_NonExistingRecord() {
	registry := NewIdentityRegistry()
	record := object.NewRecord()
	record.Set("name", object.NewString("test"))

	id, ok := registry.GetID(record)

	s.False(ok)
	s.Equal(ObjectID(0), id)
}

func (s *ObjectIdentityTestSuite) TestGetRecord_ExistingID() {
	registry := NewIdentityRegistry()
	record := object.NewRecord()
	record.Set("name", object.NewString("test"))

	id := registry.GetOrAssign(record)
	retrieved := registry.GetRecord(id)

	s.Equal(record, retrieved)
}

func (s *ObjectIdentityTestSuite) TestGetRecord_NonExistingID() {
	registry := NewIdentityRegistry()
	retrieved := registry.GetRecord(ObjectID(999))
	s.Nil(retrieved)
}

func (s *ObjectIdentityTestSuite) TestClear() {
	registry := NewIdentityRegistry()

	record := object.NewRecord()
	record.Set("name", object.NewString("test"))
	registry.GetOrAssign(record)

	s.Equal(1, registry.Count())

	registry.Clear()

	s.Equal(0, registry.Count())

	// After clear, same record should get different ID (since nextID isn't reset)
	id, ok := registry.GetID(record)
	s.False(ok)
	s.Equal(ObjectID(0), id)
}
