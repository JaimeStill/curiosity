package component_test

import (
	"reflect"
	"testing"

	"github.com/JaimeStill/curiosity/engine/ecs/component"
)

func TestIDFor_IsIdempotent(t *testing.T) {
	type T struct{}
	id1 := component.IDFor[T]()
	id2 := component.IDFor[T]()
	if id1 != id2 {
		t.Errorf("IDFor[T]() not idempotent: first = %d, second = %d", id1, id2)
	}
}

func TestIDFor_DistinctTypesGetDistinctIDs(t *testing.T) {
	type A struct{}
	type B struct{}
	a := component.IDFor[A]()
	b := component.IDFor[B]()
	if a == b {
		t.Errorf("IDFor returned the same ID for distinct types: %d", a)
	}
}

func TestIDFor_NeverReturnsInvalidID(t *testing.T) {
	type T struct{}
	if id := component.IDFor[T](); id == component.InvalidID {
		t.Error("IDFor returned InvalidID; valid IDs must be >= 1")
	}
}

func TestTypeOf_RoundTrip(t *testing.T) {
	type T struct{}
	id := component.IDFor[T]()
	got := component.TypeOf(id)
	want := reflect.TypeFor[T]()
	if got != want {
		t.Errorf("TypeOf(IDFor[T]()) = %v, want %v", got, want)
	}
}

func TestTypeOf_InvalidIDReturnsNil(t *testing.T) {
	if got := component.TypeOf(component.InvalidID); got != nil {
		t.Errorf("TypeOf(InvalidID) = %v, want nil", got)
	}
}

func TestTypeOf_OutOfRangeReturnsNil(t *testing.T) {
	if got := component.TypeOf(60000); got != nil {
		t.Errorf("TypeOf(60000) = %v, want nil", got)
	}
}

func TestSignature_SetThenHas(t *testing.T) {
	cases := []struct {
		name string
		cid  component.ID
	}{
		{"first valid CID", 1},
		{"mid word 0", 32},
		{"end of word 0", 64},
		{"start of word 1", 65},
		{"end of word 1", 128},
		{"start of word 2", 129},
		{"top of MaxCID", component.MaxCID},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s component.Signature
			s.Set(tc.cid)
			if !s.Has(tc.cid) {
				t.Errorf("after Set(%d), Has(%d) = false, want true", tc.cid, tc.cid)
			}
		})
	}
}

func TestSignature_NeighboringBitsUnset(t *testing.T) {
	cases := []struct {
		name string
		set  component.ID
		prev component.ID
		next component.ID
	}{
		{"around word boundary 64/65", 64, 63, 65},
		{"start of word 1", 65, 64, 66},
		{"mid range", 100, 99, 101},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s component.Signature
			s.Set(tc.set)
			if s.Has(tc.prev) {
				t.Errorf("Has(%d) = true after only Set(%d) — previous bit leaked", tc.prev, tc.set)
			}
			if s.Has(tc.next) {
				t.Errorf("Has(%d) = true after only Set(%d) — next bit leaked", tc.next, tc.set)
			}
		})
	}
}

func TestSignature_EmptyHasReportsFalse(t *testing.T) {
	var s component.Signature
	if s.Has(1) {
		t.Error("empty Signature.Has(1) = true, want false")
	}
}

func TestSignature_Contains(t *testing.T) {
	cases := []struct {
		name      string
		sCIDs     []component.ID
		otherCIDs []component.ID
		want      bool
	}{
		{"empty contains empty", nil, nil, true},
		{"any contains empty", []component.ID{1, 2, 3}, nil, true},
		{"empty does not contain non-empty", nil, []component.ID{1}, false},
		{"superset contains subset", []component.ID{1, 2, 3, 4}, []component.ID{2, 3}, true},
		{"disjoint sets", []component.ID{1, 2}, []component.ID{3, 4}, false},
		{"overlap but not superset", []component.ID{1, 2}, []component.ID{2, 3}, false},
		{"equal signatures", []component.ID{1, 2, 3}, []component.ID{1, 2, 3}, true},
		{"superset across word boundary", []component.ID{63, 64, 65, 66}, []component.ID{64, 65}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := component.SignatureOf(tc.sCIDs)
			other := component.SignatureOf(tc.otherCIDs)
			if got := s.Contains(other); got != tc.want {
				t.Errorf("Contains() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSignatureOf_MatchesChainedSet(t *testing.T) {
	cids := []component.ID{1, 50, 64, 65, 128, 1000}

	bulk := component.SignatureOf(cids)

	var chained component.Signature
	for _, cid := range cids {
		chained.Set(cid)
	}

	if bulk != chained {
		t.Error("SignatureOf and chained Set produced different Signatures")
	}
}
