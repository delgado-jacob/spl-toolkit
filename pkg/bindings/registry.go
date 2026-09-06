package main

import (
	"sync"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

const maxMapperID = 2_147_483_647

type mapperRegistry struct {
	mu      sync.RWMutex
	mappers map[int]*mapper.Mapper
	nextID  int64
}

func newMapperRegistry() *mapperRegistry {
	return &mapperRegistry{
		mappers: make(map[int]*mapper.Mapper),
		nextID:  1,
	}
}

func (r *mapperRegistry) add(value *mapper.Mapper) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nextID > maxMapperID {
		return 0, false
	}
	id := int(r.nextID)
	r.nextID++
	r.mappers[id] = value
	return id, true
}

func (r *mapperRegistry) get(id int) (*mapper.Mapper, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.mappers[id]
	return value, ok
}

func (r *mapperRegistry) remove(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.mappers, id)
}
