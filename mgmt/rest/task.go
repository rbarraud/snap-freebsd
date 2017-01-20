/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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

package rest

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
)

var (
	// The amount of time to buffer streaming events before flushing in seconds
	StreamingBufferWindow = 0.1

	ErrStreamingUnsupported    = errors.New("Streaming unsupported")
	ErrTaskNotFound            = errors.New("Task not found")
	ErrTaskDisabledNotRunnable = errors.New("Task is disabled. Cannot be started")
)

func (s *Server) addTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	task, err := core.CreateTaskFromContent(r.Body, nil, s.mt.CreateTask)
	if err != nil {
		respond(500, rbody.FromError(err), w)
		return
	}
	taskB := rbody.AddSchedulerTaskFromTask(task)
	taskB.Href = taskURI(r.Host, task)
	respond(201, taskB, w)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sts := s.mt.GetTasks()

	tasks := &rbody.ScheduledTaskListReturned{}
	tasks.ScheduledTasks = make([]rbody.ScheduledTask, len(sts))

	i := 0
	for _, t := range sts {
		tasks.ScheduledTasks[i] = *rbody.SchedulerTaskFromTask(t)
		tasks.ScheduledTasks[i].Href = taskURI(r.Host, t)
		i++
	}
	sort.Sort(tasks)
	respond(200, tasks, w)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	t, err1 := s.mt.GetTask(id)
	if err1 != nil {
		respond(404, rbody.FromError(err1), w)
		return
	}
	task := &rbody.ScheduledTaskReturned{}
	task.AddScheduledTask = *rbody.AddSchedulerTaskFromTask(t)
	task.Href = taskURI(r.Host, t)
	respond(200, task, w)
}

func (s *Server) watchTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.wg.Add(1)
	defer s.wg.Done()
	logger := log.WithFields(log.Fields{
		"_module": "api",
		"_block":  "watch-task",
		"client":  r.RemoteAddr,
	})

	id := p.ByName("id")

	logger.WithFields(log.Fields{
		"task-id": id,
	}).Debug("request to watch task")
	tw := &TaskWatchHandler{
		alive: true,
		mChan: make(chan rbody.StreamedTaskEvent),
	}
	tc, err1 := s.mt.WatchTask(id, tw)
	if err1 != nil {
		if strings.Contains(err1.Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromError(err1), w)
			return
		}
		respond(500, rbody.FromError(err1), w)
		return
	}

	// Make this Server Sent Events compatible
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// get a flusher type
	flusher, ok := w.(http.Flusher)
	if !ok {
		// This only works on ResponseWriters that support streaming
		respond(500, rbody.FromError(ErrStreamingUnsupported), w)
		return
	}
	// send initial stream open event
	so := rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchStreamOpen,
		Message:   "Stream opened",
	}
	fmt.Fprintf(w, "data: %s\n\n", so.ToJSON())
	flusher.Flush()

	// Get a channel for if the client notifies us it is closing the connection
	n := w.(http.CloseNotifier).CloseNotify()
	t := time.Now()
	for {
		// Write to the ResponseWriter
		select {
		case e := <-tw.mChan:
			logger.WithFields(log.Fields{
				"task-id":            id,
				"task-watcher-event": e.EventType,
			}).Debug("new event")
			switch e.EventType {
			case rbody.TaskWatchMetricEvent, rbody.TaskWatchTaskStarted:
				// The client can decide to stop receiving on the stream on Task Stopped.
				// We write the event to the buffer
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
			case rbody.TaskWatchTaskDisabled, rbody.TaskWatchTaskStopped:
				// A disabled task should end the streaming and close the connection
				fmt.Fprintf(w, "data: %s\n\n", e.ToJSON())
				// Flush since we are sending nothing new
				flusher.Flush()
				// Close out watcher removing it from the scheduler
				tc.Close()
				// exit since this client is no longer listening
				respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
			}
			// If we are at least above our minimum buffer time we flush to send
			if time.Now().Sub(t).Seconds() > StreamingBufferWindow {
				flusher.Flush()
				t = time.Now()
			}
		case <-n:
			logger.WithFields(log.Fields{
				"task-id": id,
			}).Debug("client disconnecting")
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
			return
		case <-s.killChan:
			logger.WithFields(log.Fields{
				"task-id": id,
			}).Debug("snapd exiting; disconnecting client")
			// Flush since we are sending nothing new
			flusher.Flush()
			// Close out watcher removing it from the scheduler
			tc.Close()
			// exit since this client is no longer listening
			respond(200, &rbody.ScheduledTaskWatchingEnded{}, w)
			return
		}
	}
}

func (s *Server) startTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	errs := s.mt.StartTask(id)
	if errs != nil {
		if strings.Contains(errs[0].Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromSnapErrors(errs), w)
			return
		}
		if strings.Contains(errs[0].Error(), ErrTaskDisabledNotRunnable.Error()) {
			respond(409, rbody.FromSnapErrors(errs), w)
			return
		}
		respond(500, rbody.FromSnapErrors(errs), w)
		return
	}
	// TODO should return resource
	respond(200, &rbody.ScheduledTaskStarted{ID: id}, w)
}

func (s *Server) stopTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	errs := s.mt.StopTask(id)
	if errs != nil {
		if strings.Contains(errs[0].Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromSnapErrors(errs), w)
			return
		}
		respond(500, rbody.FromSnapErrors(errs), w)
		return
	}
	respond(200, &rbody.ScheduledTaskStopped{ID: id}, w)
}

func (s *Server) removeTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	err := s.mt.RemoveTask(id)
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromError(err), w)
			return
		}
		respond(500, rbody.FromError(err), w)
		return
	}
	respond(200, &rbody.ScheduledTaskRemoved{ID: id}, w)
}

//enableTask changes the task state from Disabled to Stopped
func (s *Server) enableTask(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	tsk, err := s.mt.EnableTask(id)
	if err != nil {
		if strings.Contains(err.Error(), ErrTaskNotFound.Error()) {
			respond(404, rbody.FromError(err), w)
			return
		}
		respond(500, rbody.FromError(err), w)
		return
	}
	task := &rbody.ScheduledTaskEnabled{}
	task.AddScheduledTask = *rbody.AddSchedulerTaskFromTask(tsk)
	respond(200, task, w)
}

type TaskWatchHandler struct {
	streamCount int
	alive       bool
	mChan       chan rbody.StreamedTaskEvent
}

func (t *TaskWatchHandler) CatchCollection(m []core.Metric) {
	sm := make([]rbody.StreamedMetric, len(m))
	for i := range m {
		sm[i] = rbody.StreamedMetric{
			Namespace: m[i].Namespace().String(),
			Data:      m[i].Data(),
			Timestamp: m[i].Timestamp(),
			Tags:      m[i].Tags(),
		}
	}
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchMetricEvent,
		Message:   "",
		Event:     sm,
	}
}

func (t *TaskWatchHandler) CatchTaskStarted() {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskStarted,
	}
}

func (t *TaskWatchHandler) CatchTaskStopped() {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskStopped,
	}
}

func (t *TaskWatchHandler) CatchTaskDisabled(why string) {
	t.mChan <- rbody.StreamedTaskEvent{
		EventType: rbody.TaskWatchTaskDisabled,
		Message:   why,
	}
}

func taskURI(host string, t core.Task) string {
	return fmt.Sprintf("%s://%s/v1/tasks/%s", protocolPrefix, host, t.ID())
}
