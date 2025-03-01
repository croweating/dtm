/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"os"
	"strings"

	_ "go.uber.org/automaxprocs"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/dtm/examples"
)

var Version, Commit, Date string

var usage = `dtm is a lightweight distributed transaction manager.

usage:
    dtm [command]

Available commands:
    version           print dtm version
    dtmsvr            run dtm as a server
    dev               create all needed table and run dtm as a server
    bench             start bench server

    quick_start       run quick start example (dtm will create needed table)
    qs                same as quick_start
`

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		for name := range examples.Samples {
			fmt.Printf("%4s%-18srun a sample includes %s\n", "", name, strings.ReplaceAll(name, "_", " "))
		}
		return
	}
	if os.Args[1] == "version" {
		fmt.Printf("version: %s commit: %s built at: %s\n", Version, Commit, Date)
		return
	}
	logger.Infof("starting dtm....")
	common.MustLoadConfig()
	if common.Config.ExamplesDB.Driver != "" {
		dtmcli.SetCurrentDBType(common.Config.ExamplesDB.Driver)
	}
	if os.Args[1] != "dtmsvr" { // 实际线上运行，只启动dtmsvr，不准备table相关的数据
		registry.WaitStoreUp()
		dtmsvr.PopulateDB(true)
		examples.PopulateDB(true)
	}
	dtmsvr.StartSvr()              // 启动dtmsvr的api服务
	go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询

	switch os.Args[1] {
	case "quick_start", "qs":
		// quick_start 比较独立，单独作为一个例子运行，方便新人上手
		examples.QsStartSvr()
		examples.QsFireRequest()
	case "dev", "dtmsvr": // do nothing, not fallthrough
	default:
		// 下面是各类的例子
		examples.GrpcStartup()
		examples.BaseAppStartup()

		sample := examples.Samples[os.Args[1]]
		logger.FatalfIf(sample == nil, "no sample name for %s", os.Args[1])
		sample.Action()
	}
	select {}
}
