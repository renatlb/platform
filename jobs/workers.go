// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package jobs

import (
	"sync"

	l4g "github.com/alecthomas/log4go"
	ejobs "github.com/mattermost/platform/einterfaces/jobs"
	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
)

type Workers struct {
	startOnce     sync.Once
	watcher       *Watcher

	DataRetention model.Worker
	// SearchIndexing model.Job

	listenerId    string
}

func InitWorkers() *Workers {
	workers := &Workers{
	// 	SearchIndexing: MakeTestJob(s, "SearchIndexing"),
	}
	workers.watcher = MakeWatcher(workers)

	if dataRetentionInterface := ejobs.GetDataRetentionInterface(); dataRetentionInterface != nil {
		workers.DataRetention = dataRetentionInterface.MakeWorker()
	}

	return workers
}

func (workers *Workers) Start() *Workers {
	l4g.Info("Starting workers")

	workers.startOnce.Do(func() {
		if workers.DataRetention != nil && *utils.Cfg.DataRetentionSettings.Enable {
			go workers.DataRetention.Run()
		}

		// go workers.SearchIndexing.Run()

		go workers.watcher.Start()
	})

	workers.listenerId = utils.AddConfigListener(workers.handleConfigChange)

	return workers
}

func (workers *Workers) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	if workers.DataRetention != nil {
		if !*oldConfig.DataRetentionSettings.Enable && *newConfig.DataRetentionSettings.Enable {
			go workers.DataRetention.Run()
		} else if *oldConfig.DataRetentionSettings.Enable && !*newConfig.DataRetentionSettings.Enable {
			workers.DataRetention.Stop()
		}
	}
}

func (workers *Workers) Stop() *Workers {
	utils.RemoveConfigListener(workers.listenerId)

	workers.watcher.Stop()

	if workers.DataRetention != nil && *utils.Cfg.DataRetentionSettings.Enable {
		workers.DataRetention.Stop()
	}
	// workers.SearchIndexing.Stop()

	l4g.Info("Stopped workers")

	return workers
}
