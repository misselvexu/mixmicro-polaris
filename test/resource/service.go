/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package resource

import (
	"fmt"
	"time"

	api "github.com/polarismesh/polaris-server/common/api/v1"
	"github.com/polarismesh/polaris-server/common/utils"
)

const (
	serviceName = "test-service-%v-%v"
)

// CreateServices creates services
func CreateServices(namespace *api.Namespace) []*api.Service {
	nowStr := time.Now().Format("2006-01-02T15:04:05.000000")

	var services []*api.Service
	for index := 0; index < 2; index++ {
		name := fmt.Sprintf(serviceName, nowStr, index)

		service := &api.Service{
			Name:       utils.NewStringValue(name),
			Namespace:  namespace.GetName(),
			Metadata:   map[string]string{"test": "test"},
			Ports:      utils.NewStringValue("8,8"),
			Business:   utils.NewStringValue("test"),
			Department: utils.NewStringValue("test"),
			CmdbMod1:   utils.NewStringValue("test"),
			CmdbMod2:   utils.NewStringValue("test"),
			CmdbMod3:   utils.NewStringValue("test"),
			Comment:    utils.NewStringValue("test"),
			Owners:     utils.NewStringValue("test"),
		}
		services = append(services, service)
	}

	return services
}

// UpdateServices 更新测试服务
func UpdateServices(services []*api.Service) {
	for _, service := range services {
		service.Metadata = map[string]string{"update": "update"}
		service.Ports = utils.NewStringValue("4,4")
		service.Business = utils.NewStringValue("update")
		service.Department = utils.NewStringValue("update")
		service.CmdbMod1 = utils.NewStringValue("update")
		service.CmdbMod2 = utils.NewStringValue("update")
		service.CmdbMod3 = utils.NewStringValue("update")
		service.Comment = utils.NewStringValue("update")
		service.Owners = utils.NewStringValue("update")
	}
}
