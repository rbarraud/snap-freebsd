// +build legacy small medium large

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fixtures

import (
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

var taskCatalog map[string]core.Task = map[string]core.Task{
	"Task1": &mockTask{
		MyID:                "qwertyuiop",
		MyName:              "TASK1.0",
		MyDeadline:          "4",
		MyCreationTimestamp: time.Now().Unix(),
		MyLastRunTimestamp:  time.Now().Unix(),
		MyHitCount:          44,
		MyMissCount:         8,
		MyState:             "failed",
		MyHref:              "http://localhost:8181/v2/tasks/qwertyuiop"},
	"Task2": &mockTask{
		MyID:                "asdfghjkl",
		MyName:              "TASK2.0",
		MyDeadline:          "4",
		MyCreationTimestamp: time.Now().Unix(),
		MyLastRunTimestamp:  time.Now().Unix(),
		MyHitCount:          33,
		MyMissCount:         7,
		MyState:             "passed",
		MyHref:              "http://localhost:8181/v2/tasks/asdfghjkl"}}

type mockTask struct {
	MyID                 string            `json:"id"`
	MyName               string            `json:"name"`
	MyDeadline           string            `json:"deadline"`
	MyWorkflow           *wmap.WorkflowMap `json:"workflow,omitempty"`
	MySchedule           *core.Schedule    `json:"schedule,omitempty"`
	MyCreationTimestamp  int64             `json:"creation_timestamp,omitempty"`
	MyLastRunTimestamp   int64             `json:"last_run_timestamp,omitempty"`
	MyHitCount           int               `json:"hit_count,omitempty"`
	MyMissCount          int               `json:"miss_count,omitempty"`
	MyFailedCount        int               `json:"failed_count,omitempty"`
	MyLastFailureMessage string            `json:"last_failure_message,omitempty"`
	MyState              string            `json:"task_state"`
	MyHref               string            `json:"href"`
}

func (t *mockTask) ID() string                        { return t.MyID }
func (t *mockTask) State() core.TaskState             { return core.TaskSpinning }
func (t *mockTask) HitCount() uint                    { return 0 }
func (t *mockTask) GetName() string                   { return t.MyName }
func (t *mockTask) SetName(string)                    { return }
func (t *mockTask) SetID(string)                      { return }
func (t *mockTask) MissedCount() uint                 { return 0 }
func (t *mockTask) FailedCount() uint                 { return 0 }
func (t *mockTask) LastFailureMessage() string        { return "" }
func (t *mockTask) LastRunTime() *time.Time           { return &time.Time{} }
func (t *mockTask) CreationTime() *time.Time          { return &time.Time{} }
func (t *mockTask) DeadlineDuration() time.Duration   { return 4 }
func (t *mockTask) SetDeadlineDuration(time.Duration) { return }
func (t *mockTask) SetTaskID(id string)               { return }
func (t *mockTask) SetStopOnFailure(int)              { return }
func (t *mockTask) GetStopOnFailure() int             { return 0 }
func (t *mockTask) Option(...core.TaskOption) core.TaskOption {
	return core.TaskDeadlineDuration(0)
}
func (t *mockTask) WMap() *wmap.WorkflowMap {
	return wmap.NewWorkflowMap()
}
func (t *mockTask) Schedule() schedule.Schedule {
	return schedule.NewSimpleSchedule(time.Second * 1)
}
func (t *mockTask) MaxFailures() int { return 10 }

type MockTaskManager struct{}

func (m *MockTaskManager) GetTask(id string) (core.Task, error) {
	href := "http://localhost:8181/v2/tasks/" + id
	return &mockTask{
		MyID:                id,
		MyName:              "NewTaskCreated",
		MyCreationTimestamp: time.Now().Unix(),
		MyLastRunTimestamp:  time.Now().Unix(),
		MyHitCount:          22,
		MyMissCount:         4,
		MyState:             "failed",
		MyHref:              href}, nil
}
func (m *MockTaskManager) CreateTask(
	sch schedule.Schedule,
	wmap *wmap.WorkflowMap,
	start bool,
	opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	return &mockTask{
		MyID:                "MyTaskID",
		MyName:              "NewTaskCreated",
		MySchedule:          &core.Schedule{},
		MyCreationTimestamp: time.Now().Unix(),
		MyLastRunTimestamp:  time.Now().Unix(),
		MyHitCount:          99,
		MyMissCount:         5,
		MyState:             "failed",
		MyHref:              "http://localhost:8181/v2/tasks/MyTaskID"}, nil
}
func (m *MockTaskManager) GetTasks() map[string]core.Task {
	return taskCatalog
}
func (m *MockTaskManager) StartTask(id string) []serror.SnapError { return nil }
func (m *MockTaskManager) StopTask(id string) []serror.SnapError  { return nil }
func (m *MockTaskManager) RemoveTask(id string) error             { return nil }
func (m *MockTaskManager) WatchTask(id string, handler core.TaskWatcherHandler) (core.TaskWatcherCloser, error) {
	return nil, nil
}
func (m *MockTaskManager) EnableTask(id string) (core.Task, error) {
	return &mockTask{
		MyID:                "alskdjf",
		MyName:              "Task2",
		MyCreationTimestamp: time.Now().Unix(),
		MyLastRunTimestamp:  time.Now().Unix(),
		MyHitCount:          44,
		MyMissCount:         8,
		MyState:             "failed",
		MyHref:              "http://localhost:8181/v2/tasks/alskdjf"}, nil
}

// Mock task used in the 'Add tasks' test in rest_v1_test.go
const TASK = `{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "max-failures": 10,
    "workflow": {
        "collect": {
            "metrics": {
                "/one/two/three": {}
            }
        }
    }
}
`

// These constants are the expected responses from running the task tests in
// rest_v1_test.go on the task routes found in mgmt/rest/server.go
const (
	GET_TASKS_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Scheduled tasks retrieved",
    "type": "scheduled_task_list_returned",
    "version": 1
  },
  "body": {
    "ScheduledTasks": [
      {
        "id": "qwertyuiop",
        "name": "TASK1.0",
        "deadline": "4ns",
        "creation_timestamp": -62135596800,
        "last_run_timestamp": -1,
        "task_state": "Running",
        "href": "http://localhost:%d/v1/tasks/qwertyuiop"
      },
      {
        "id": "asdfghjkl",
        "name": "TASK2.0",
        "deadline": "4ns",
        "creation_timestamp": -62135596800,
        "last_run_timestamp": -1,
        "task_state": "Running",
        "href": "http://localhost:%d/v1/tasks/asdfghjkl"
      }
    ]
  }
}`

	GET_TASKS_RESPONSE2 = `{
  "meta": {
    "code": 200,
    "message": "Scheduled tasks retrieved",
    "type": "scheduled_task_list_returned",
    "version": 1
  },
  "body": {
    "ScheduledTasks": [
      {
        "id": "asdfghjkl",
        "name": "TASK2.0",
        "deadline": "4ns",
        "creation_timestamp": -62135596800,
        "last_run_timestamp": -1,
        "task_state": "Running",
        "href": "http://localhost:%d/v1/tasks/asdfghjkl"
      },
      {
        "id": "qwertyuiop",
        "name": "TASK1.0",
        "deadline": "4ns",
        "creation_timestamp": -62135596800,
        "last_run_timestamp": -1,
        "task_state": "Running",
        "href": "http://localhost:%d/v1/tasks/qwertyuiop"
      }
    ]
  }
}`

	GET_TASK_RESPONSE = `{
  "meta": {
    "code": 200,
    "message": "Scheduled task (:1234) returned",
    "type": "scheduled_task_returned",
    "version": 1
  },
  "body": {
    "id": ":1234",
    "name": "NewTaskCreated",
    "deadline": "4ns",
    "workflow": {
      "collect": {
        "metrics": {}
      }
    },
    "schedule": {
      "type": "simple",
      "interval": "1s"
    },
    "creation_timestamp": -62135596800,
    "last_run_timestamp": -1,
    "task_state": "Running",
    "href": "http://localhost:%d/v1/tasks/:1234"
  }
}`

	ADD_TASK_RESPONSE = `{
  "meta": {
    "code": 201,
    "message": "Scheduled task created (MyTaskID)",
    "type": "scheduled_task_created",
    "version": 1
  },
  "body": {
    "id": "MyTaskID",
    "name": "NewTaskCreated",
    "deadline": "4ns",
    "workflow": {
      "collect": {
        "metrics": {}
      }
    },
    "schedule": {
      "type": "simple",
      "interval": "1s"
    },
    "creation_timestamp": -62135596800,
    "last_run_timestamp": -1,
    "task_state": "Running",
    "href": "http://localhost:%d/v1/tasks/MyTaskID"
  }
}`

	START_TASK_RESPONSE_ID_START = `{
  "meta": {
    "code": 200,
    "message": "Scheduled task (MockTask1234) started",
    "type": "scheduled_task_started",
    "version": 1
  },
  "body": {
    "id": "MockTask1234"
  }
}`

	STOP_TASK_RESPONSE_ID_STOP = `{
  "meta": {
    "code": 200,
    "message": "Scheduled task (MockTask1234) stopped",
    "type": "scheduled_task_stopped",
    "version": 1
  },
  "body": {
    "id": "MockTask1234"
  }
}`

	ENABLE_TASK_RESPONSE_ID_ENABLE = `{
  "meta": {
    "code": 200,
    "message": "Disabled task (alskdjf) enabled",
    "type": "scheduled_task_enabled",
    "version": 1
  },
  "body": {
    "id": "alskdjf",
    "name": "Task2",
    "deadline": "4ns",
    "workflow": {
      "collect": {
        "metrics": {}
      }
    },
    "schedule": {
      "type": "simple",
      "interval": "1s"
    },
    "creation_timestamp": -62135596800,
    "last_run_timestamp": -1,
    "task_state": "Running",
    "href": ""
  }
}`

	REMOVE_TASK_RESPONSE_ID = `{
  "meta": {
    "code": 200,
    "message": "Scheduled task (MockTask1234) removed",
    "type": "scheduled_task_removed",
    "version": 1
  },
  "body": {
    "id": "MockTask1234"
  }
}`
)
