package pgx

import (
	"reflect"

	"testing"
)

type emptyStruct struct{}

type unserializableStruct struct {
	Lorem func()
}

type simpleStruct struct {
	Value string
}

func TestSerializeNil(t *testing.T) {
	bytes := serialize[emptyStruct](nil)

	if len(bytes) != 0 {
		t.Fatalf("Expected to not empty bytes array, but got %v", bytes)
	}
}

func TestUnserializable(t *testing.T) {
	defer assertPanic(t, "Could not encode data.")

	impossible := &unserializableStruct{
		func() { panic("does not work") },
	}
	serialize(impossible)
}

func TestSerialize(t *testing.T) {
	original := &simpleStruct{"wow!"}
	bytes := serialize(original)

	if len(bytes) == 0 {
		t.Fatalf("Expected to have a not empty bytes array, but got an empty array.")
	}

	deserialized := deserialize[simpleStruct](bytes)
	if !reflect.DeepEqual(original, deserialized) {
		t.Fatalf(
			"Expected original object to be equal to its deserialized counterpart, but %v != %v.",
			original,
			deserialized,
		)
	}
}

func TestDeserializeEmpty(t *testing.T) {
	deserialized := deserialize[simpleStruct]([]byte{})

	if deserialized != nil {
		t.Fatalf("Expected to get nil on empty bytes, but got %v", deserialized)
	}
}

func TestUndeserializable(t *testing.T) {
	defer assertPanic(t, "Could not decode data.")

	deserialize[unserializableStruct]([]byte{'A'})
}

func assertPanic(t *testing.T, message string) {
	recovery := recover()

	if recovery == nil {
		t.Fatalf("Expected panic with \"%s\", but didn't", message)
	}

	panicMessage, ok := recovery.(string)

	if !ok {
		t.Fatalf("Expected panic with \"%s\", but got %v", message, recovery)
	}

	if panicMessage != message {
		t.Fatalf("Expected panic with \"%s\", but got \"%s\"", message, panicMessage)
	}
}
