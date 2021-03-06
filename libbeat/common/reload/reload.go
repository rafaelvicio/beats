// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package reload

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/elastic/beats/libbeat/common"
)

// Register holds a registry of reloadable objects
var Register = newRegistry()

// ConfigWithMeta holds a pair of common.Config and optional metadata for it
type ConfigWithMeta struct {
	// Config to store
	Config *common.Config

	// Meta data related to this config
	Meta *common.MapStrPointer
}

// ReloadableList provides a method to reload the configuration of a list of entities
type ReloadableList interface {
	Reload(configs []*ConfigWithMeta) error
}

// Reloadable provides a method to reload the configuration of an entity
type Reloadable interface {
	Reload(config *ConfigWithMeta) error
}

type registry struct {
	sync.RWMutex
	confsLists map[string]ReloadableList
	confs      map[string]Reloadable
}

func newRegistry() *registry {
	return &registry{
		confsLists: make(map[string]ReloadableList),
		confs:      make(map[string]Reloadable),
	}
}

// Register declares a reloadable object
func (r *registry) Register(name string, obj Reloadable) error {
	r.Lock()
	defer r.Unlock()

	if obj == nil {
		return errors.New("got a nil object")
	}

	if r.nameTaken(name) {
		return errors.Errorf("%s configuration list is already registered", name)
	}

	r.confs[name] = obj
	return nil
}

// RegisterList declares a reloadable list of configurations
func (r *registry) RegisterList(name string, list ReloadableList) error {
	r.Lock()
	defer r.Unlock()

	if list == nil {
		return errors.New("got a nil object")
	}

	if r.nameTaken(name) {
		return errors.Errorf("%s configuration is already registered", name)
	}

	r.confsLists[name] = list
	return nil
}

// MustRegister declares a reloadable object
func (r *registry) MustRegister(name string, obj Reloadable) {
	if err := r.Register(name, obj); err != nil {
		panic(err)
	}
}

// MustRegisterList declares a reloadable object list
func (r *registry) MustRegisterList(name string, list ReloadableList) {
	if err := r.RegisterList(name, list); err != nil {
		panic(err)
	}
}

// Get returns the reloadable object with the given name, nil if not found
func (r *registry) Get(name string) Reloadable {
	r.RLock()
	defer r.RUnlock()
	return r.confs[name]
}

// GetList returns the reloadable list with the given name, nil if not found
func (r *registry) GetList(name string) ReloadableList {
	r.RLock()
	defer r.RUnlock()
	return r.confsLists[name]
}

func (r *registry) nameTaken(name string) bool {
	if _, ok := r.confs[name]; ok {
		return true
	}

	if _, ok := r.confsLists[name]; ok {
		return true
	}

	return false
}
