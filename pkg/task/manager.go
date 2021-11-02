// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"sort"
	"sync"
	"time"
)

const taskCapacity = 100

const StatusProcessing = "PROCESSING"
const StatusCompleted = "COMPLETED"
const StatusFailed = "FAILED"

type Task struct {
	ID       int        `json:"id"`
	Created  time.Time  `json:"created"`
	Finished *time.Time `json:"finished"`
	Source   string     `json:"source"`
	Status   string     `json:"status"`
	Reason   string     `json:"reason"`
}

type manager struct {
	id    int
	mutex sync.Mutex
	tasks []*Task
}

var Manager *manager

func init() {
	Manager = &manager{
		mutex: sync.Mutex{},
		id:    0,
		// Once the number of tasks in the task list exceeds the
		// capacity value, the earlier created task items are
		// overwritten by the newly created ones.
		tasks: make([]*Task, taskCapacity),
	}
}

func (m *manager) Create(source string) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.id >= len(m.tasks) {
		m.id = 0
	}

	id := m.id
	m.tasks[m.id] = &Task{
		ID:       m.id,
		Created:  time.Now(),
		Finished: nil,
		Source:   source,
		Status:   StatusProcessing,
		Reason:   "",
	}
	m.id++

	return id
}

func (m *manager) Finish(id int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task := m.tasks[id]
	if task != nil {
		if err != nil {
			task.Status = StatusFailed
			task.Reason = err.Error()
		} else {
			task.Status = StatusCompleted
		}
		now := time.Now()
		task.Finished = &now
	}
}

func (m *manager) List() []*Task {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tasks := make([]*Task, 0)
	for _, task := range m.tasks {
		if task != nil {
			tasks = append(tasks, task)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Created.After(tasks[j].Created)
	})

	return tasks
}
