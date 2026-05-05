package domain

import "encoding/json"

type Nullable[T any] struct {
	Value T
	Valid bool
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Valid = true
	return json.Unmarshal(data, &n.Value)
}