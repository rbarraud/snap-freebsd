/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package scheduler

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var (
	flSchedulerQueueSize = cli.StringFlag{
		Name:   "work-manager-queue-size",
		Usage:  fmt.Sprintf("Size of the work manager queue (default: %v)", defaultWorkManagerQueueSize),
		EnvVar: "WORK_MANAGER_QUEUE_SIZE",
	}

	flSchedulerPoolSize = cli.StringFlag{
		Name:   "work-manager-pool-size",
		Usage:  fmt.Sprintf("Size of the work manager pool (default: %v)", defaultWorkManagerPoolSize),
		EnvVar: "WORK_MANAGER_POOL_SIZE",
	}

	// Flags consumed by snapd
	Flags = []cli.Flag{flSchedulerQueueSize, flSchedulerPoolSize}
)
