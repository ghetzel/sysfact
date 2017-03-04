package main

import (
	"github.com/ghetzel/go-stockutil/maputil"
	"strings"
)

type Tuple struct {
	Key           string
	Value         interface{}
	NormalizedKey string
	Tags          map[string]interface{}
}

type TupleSet []Tuple

func (self TupleSet) Remove(i int) {
	self = append(self[:i], self[i+1:]...)
}

func (self TupleSet) Len() int {
	return len(self)
}

func (self TupleSet) Less(i, j int) bool {
	a := self[i]
	b := self[j]

	return (strings.Compare(a.Key, b.Key) < 0)
}

func (self TupleSet) Swap(i, j int) {
	tmp := self[i]
	self[i] = self[j]
	self[j] = tmp
}

func (self TupleSet) ToMap(flat bool) map[string]interface{} {
	output := make(map[string]interface{})

	for _, tuple := range self {
		output[tuple.Key] = tuple.Value
	}

	if !flat {
		if nestedOutput, err := maputil.DiffuseMap(output, `.`); err == nil {
			output = nestedOutput
		}
	}

	return output
}
