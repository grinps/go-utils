package config

import (
	"context"
)

var InvalidKey = simpleStringKey(InvalidKeyName)

type simpleStringKey string

func (key simpleStringKey) Key() Key {
	return key
}

func (key simpleStringKey) Name(ctx context.Context) string {
	if key == "" {
		return InvalidKeyName
	}
	return string(key)
}

func (key simpleStringKey) HasChild(ctx context.Context) bool {
	return false
}
func (key simpleStringKey) Child(ctx context.Context) Key {
	return InvalidKey
}

type ComplexKey []Key

func (key *ComplexKey) Key() Key {
	return key
}

func (key *ComplexKey) HasChild(ctx context.Context) bool {
	if key != nil {
		var asAnySlice = []Key(*key)
		if len(asAnySlice) > 1 {
			return true
		}
	}
	return false
}

func (key *ComplexKey) Child(ctx context.Context) Key {
	if key != nil {
		var asAnySlice = []Key(*key)
		sliceLength := len(asAnySlice)
		switch {
		case sliceLength > 1:
			childKey := ComplexKey(asAnySlice[1:])
			return &childKey
		case sliceLength == 1:
			return key
		case sliceLength == 0:
			return InvalidKey
		}
	}
	return InvalidKey
}

func (key *ComplexKey) Name(ctx context.Context) string {
	if key != nil {
		var asAnySlice = []Key(*key)
		sliceLength := len(asAnySlice)
		switch {
		case sliceLength >= 1:
			key := asAnySlice[0]
			switch key.(type) {
			case SimpleKey:
				return key.(SimpleKey).Name(ctx)
			default:
				return InvalidKeyName
			}
		default:
			return InvalidKeyName
		}
	}
	return InvalidKeyName
}
