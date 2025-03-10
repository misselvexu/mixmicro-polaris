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

package test

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	api "github.com/polarismesh/polaris-server/common/api/v1"
	"github.com/polarismesh/polaris-server/common/utils"
	"github.com/polarismesh/polaris-server/service"
)

// 测试新增服务
func TestCreateService(t *testing.T) {
	t.Run("正常创建服务", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 9)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		if serviceResp.GetName().GetValue() == serviceReq.GetName().GetValue() &&
			serviceResp.GetNamespace().GetValue() == serviceReq.GetNamespace().GetValue() &&
			serviceResp.GetToken().GetValue() != "" {
			t.Logf("pass")
		} else {
			t.Fatalf("error: %+v", serviceResp)
		}
	})

	t.Run("创建重复名字的服务，会返回失败", func(t *testing.T) {
		serviceReq, _ := createCommonService(t, 9)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		resp := server.CreateService(defaultCtx, serviceReq)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("创建服务，删除，再次创建，可以正常创建", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 100)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		req := &api.Service{
			Name:      utils.NewStringValue(serviceResp.GetName().GetValue()),
			Namespace: utils.NewStringValue(serviceResp.GetNamespace().GetValue()),
			Token:     utils.NewStringValue(serviceResp.GetToken().GetValue()),
		}
		removeCommonServices(t, []*api.Service{req})

		if resp := server.CreateService(defaultCtx, serviceReq); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}

		t.Logf("pass")
	})
	t.Run("并发创建服务", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 500; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				serviceReq, _ := createCommonService(t, index)
				cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
			}(i)
		}
		wg.Wait()
	})
	t.Run("命名空间不存在，无法创建服务", func(t *testing.T) {
		service := &api.Service{
			Name:      utils.NewStringValue("abc"),
			Namespace: utils.NewStringValue("123456"),
			Owners:    utils.NewStringValue("my"),
		}
		resp := server.CreateService(defaultCtx, service)
		if respSuccess(resp) {
			t.Fatalf("error")
		}
		t.Logf("pass: %s", resp.GetInfo().GetValue())
	})
	t.Run("创建服务，metadata个数太多，报错", func(t *testing.T) {
		svc := &api.Service{
			Name:      utils.NewStringValue("999"),
			Namespace: utils.NewStringValue("Polaris"),
			Owners:    utils.NewStringValue("my"),
		}
		svc.Metadata = make(map[string]string)
		for i := 0; i < service.MaxMetadataLength+1; i++ {
			svc.Metadata[fmt.Sprintf("aa-%d", i)] = "value"
		}
		if resp := server.CreateService(defaultCtx, svc); !respSuccess(resp) {
			t.Logf("%s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})
}

// delete services
func TestRemoveServices(t *testing.T) {
	t.Run("删除单个服务，删除成功", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 59)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		req := &api.Service{
			Name:      utils.NewStringValue(serviceResp.GetName().GetValue()),
			Namespace: utils.NewStringValue(serviceResp.GetNamespace().GetValue()),
			Token:     utils.NewStringValue(serviceResp.GetToken().GetValue()),
		}

		// wait for data cache
		time.Sleep(time.Second * 2)
		removeCommonServices(t, []*api.Service{req})
		out := server.GetServices(context.Background(), map[string]string{"name": req.GetName().GetValue()})
		if !respSuccess(out) {
			t.Fatalf(out.GetInfo().GetValue())
		}
		if len(out.GetServices()) != 0 {
			t.Fatalf("error: %d", len(out.GetServices()))
		}
	})

	t.Run("删除多个服务，删除成功", func(t *testing.T) {
		var reqs []*api.Service
		for i := 0; i < 100; i++ {
			serviceReq, serviceResp := createCommonService(t, i)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
			req := &api.Service{
				Name:      utils.NewStringValue(serviceResp.GetName().GetValue()),
				Namespace: utils.NewStringValue(serviceResp.GetNamespace().GetValue()),
				Token:     utils.NewStringValue(serviceResp.GetToken().GetValue()),
			}
			reqs = append(reqs, req)
		}

		// wait for data cache
		time.Sleep(time.Second * 2)
		removeCommonServices(t, reqs)
	})

	t.Run("创建一个服务，马上删除，可以正常删除", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 19)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		req := &api.Service{
			Name:      utils.NewStringValue(serviceResp.GetName().GetValue()),
			Namespace: utils.NewStringValue(serviceResp.GetNamespace().GetValue()),
			Token:     utils.NewStringValue(serviceResp.GetToken().GetValue()),
		}
		removeCommonServices(t, []*api.Service{req})
	})
	t.Run("创建服务和实例，删除服务，删除失败", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 19)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		_, instanceResp := createCommonInstance(t, serviceResp, 100)
		defer cleanInstance(instanceResp.GetId().GetValue())

		resp := server.DeleteService(defaultCtx, serviceResp)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})

	t.Run("并发删除服务", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			serviceReq, serviceResp := createCommonService(t, i)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
			req := &api.Service{
				Name:      utils.NewStringValue(serviceResp.GetName().GetValue()),
				Namespace: utils.NewStringValue(serviceResp.GetNamespace().GetValue()),
				Token:     utils.NewStringValue(serviceResp.GetToken().GetValue()),
			}

			wg.Add(1)
			go func(reqs []*api.Service) {
				defer wg.Done()
				removeCommonServices(t, reqs)
			}([]*api.Service{req})
		}
		wg.Wait()
	})
}

// 关联测试
func TestDeleteService2(t *testing.T) {
	t.Run("存在路由配置的情况下，删除服务会失败", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 20)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		// 创建一个服务配置
		createCommonRoutingConfig(t, serviceResp, 10, 10)
		defer cleanCommonRoutingConfig(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
		// 删除服务
		resp := server.DeleteService(defaultCtx, serviceResp)
		if respSuccess(resp) {
			t.Fatalf("error")
		}
		t.Logf("pass: %s", resp.GetInfo().GetValue())
	})
	t.Run("重复删除服务，返回成功", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 20)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		removeCommonServices(t, []*api.Service{serviceResp})
		removeCommonServices(t, []*api.Service{serviceResp})
	})
	t.Run("存在别名的情况下，删除服务会失败", func(t *testing.T) {
		serviceReq, serviceResp := createCommonService(t, 20)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		aliasResp1 := createCommonAlias(serviceResp, "", defaultAliasNs, api.AliasType_CL5SID)
		defer cleanServiceName(aliasResp1.Alias.Alias.Value, serviceResp.Namespace.Value)
		aliasResp2 := createCommonAlias(serviceResp, "", defaultAliasNs, api.AliasType_CL5SID)
		defer cleanServiceName(aliasResp2.Alias.Alias.Value, serviceResp.Namespace.Value)

		// 删除服务
		resp := server.DeleteService(defaultCtx, serviceResp)
		if respSuccess(resp) {
			t.Fatalf("error")
		}
		t.Logf("pass: %s", resp.GetInfo().GetValue())
	})
}

// 测试批量获取服务负责人
func TestGetServiceOwner(t *testing.T) {
	t.Run("服务个数为0，返回错误", func(t *testing.T) {
		var reqs []*api.Service
		if resp := server.GetServiceOwner(defaultCtx, reqs); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
	})

	t.Run("服务个数超过100，返回错误", func(t *testing.T) {
		reqs := make([]*api.Service, 0, 101)
		for i := 0; i < 101; i++ {
			req := &api.Service{
				Namespace: utils.NewStringValue("Test"),
				Name:      utils.NewStringValue("test"),
			}
			reqs = append(reqs, req)
		}
		if resp := server.GetServiceOwner(defaultCtx, reqs); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
	})

	t.Run("查询100个超长服务名的服务负责人，数据库不会报错", func(t *testing.T) {
		reqs := make([]*api.Service, 0, 100)
		for i := 0; i < 100; i++ {
			req := &api.Service{
				Namespace: utils.NewStringValue("Development"),
				Name:      utils.NewStringValue(genSpecialStr(128)),
			}
			reqs = append(reqs, req)
		}
		if resp := server.GetServiceOwner(defaultCtx, reqs); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		t.Log("pass")
	})
}

// 测试获取服务函数
func TestGetService(t *testing.T) {
	t.Run("查询服务列表，可以正常返回", func(t *testing.T) {
		resp := server.GetServices(context.Background(), map[string]string{})
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}
	})
	t.Run("查询服务列表，只有limit和offset，可以正常返回预计个数的服务", func(t *testing.T) {
		total := 20
		reqs := make([]*api.Service, 0, total)
		for i := 0; i < total; i++ {
			serviceReq, _ := createCommonService(t, i+10)
			reqs = append(reqs, serviceReq)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
		}

		// 创建完，直接查询
		filters := map[string]string{"offset": "0", "limit": "100"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}

		if resp.GetSize().GetValue() >= uint32(total) && resp.GetSize().GetValue() <= 100 {
			t.Logf("pass")
		} else {
			t.Fatalf("error: %d %d", resp.GetSize().GetValue(), total)
		}
	})

	t.Run("查询服务列表，没有filter，只回复默认的service", func(t *testing.T) {
		total := 10
		for i := 0; i < total; i++ {
			serviceReq, _ := createCommonService(t, i+10)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
		}

		resp := server.GetServices(context.Background(), map[string]string{})
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}
		if resp.GetSize().GetValue() >= 10 {
			t.Logf("pass")
		} else {
			t.Fatalf("error: %d", resp.GetSize().GetValue())
		}
	})
	t.Run("查询服务列表，只能查询到源服务，无法查询到别名", func(t *testing.T) {
		total := 10
		for i := 0; i < total; i++ {
			_, serviceResp := createCommonService(t, i+102)
			defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
			aliasResp := createCommonAlias(serviceResp, "", defaultAliasNs, api.AliasType_CL5SID)
			defer cleanServiceName(aliasResp.Alias.Alias.Value, serviceResp.Namespace.Value)
		}
		resp := server.GetServices(context.Background(), map[string]string{"owner": "service-owner-102"})
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}
		if resp.GetSize().GetValue() != 1 {
			t.Fatalf("error: %d", resp.GetSize().GetValue())
		}
	})
}

// 测试获取服务列表，参数校验
func TestGetServices2(t *testing.T) {
	t.Run("查询服务列表，limit有最大为100的限制", func(t *testing.T) {
		total := 101
		for i := 0; i < total; i++ {
			serviceReq, _ := createCommonService(t, i+10)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
		}

		filters := map[string]string{"offset": "0", "limit": "600"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}
		if resp.GetSize().GetValue() == service.QueryMaxLimit {
			t.Logf("pass")
		} else {
			t.Fatalf("error: %d", resp.GetSize().GetValue())
		}
	})
	t.Run("查询服务列表，offset参数不为int，返回错误", func(t *testing.T) {
		filters := map[string]string{"offset": "abc", "limit": "200"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("查询服务列表，limit参数不为int，返回错误", func(t *testing.T) {
		filters := map[string]string{"offset": "0", "limit": "ss"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("查询服务列表，offset参数为负数，返回错误", func(t *testing.T) {
		filters := map[string]string{"offset": "-100", "limit": "10"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("查询服务列表，limit参数为负数，返回错误", func(t *testing.T) {
		filters := map[string]string{"offset": "100", "limit": "-10"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("查询服务列表，单独提供port参数，返回错误", func(t *testing.T) {
		filters := map[string]string{"port": "100"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("查询服务列表，port参数有误，返回错误", func(t *testing.T) {
		filters := map[string]string{"port": "p100", "host": "127.0.0.1"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Logf("pass: %s", resp.Info.GetValue())
		} else {
			t.Fatalf("error")
		}
	})
}

// 有基础的过滤条件的查询服务列表
func TestGetService3(t *testing.T) {
	t.Run("根据服务名，可以正常过滤", func(t *testing.T) {
		var reqs []*api.Service
		serviceReq, _ := createCommonService(t, 100)
		reqs = append(reqs, serviceReq)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		namespaceReq, _ := createCommonNamespace(t, 100)
		defer cleanNamespace(namespaceReq.GetName().GetValue())

		serviceReq.Namespace = utils.NewStringValue(namespaceReq.GetName().GetValue())
		if resp := server.CreateService(defaultCtx, serviceReq); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		reqs = append(reqs, serviceReq)
		defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())

		name := serviceReq.GetName().GetValue()
		filters := map[string]string{"offset": "0", "limit": "10", "name": name}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}

		CheckGetService(t, reqs, resp.GetServices())
		t.Logf("pass")
	})

	t.Run("多重过滤条件，可以生效", func(t *testing.T) {
		total := 10
		var name, namespace string
		for i := 0; i < total; i++ {
			serviceReq, _ := createCommonService(t, 100)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
			if i == 5 {
				name = serviceReq.GetName().GetValue()
				namespace = serviceReq.GetNamespace().GetValue()
			}
		}
		filters := map[string]string{"offset": "0", "limit": "10", "name": name, "namespace": namespace}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if len(resp.Services) != 1 {
			t.Fatalf("error: %d", len(resp.Services))
		}
	})

	t.Run("owner过滤条件会生效", func(t *testing.T) {
		total := 60
		for i := 0; i < total; i++ {
			serviceReq, _ := createCommonService(t, i+10)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
		}

		filters := map[string]string{"offset": "0", "limit": "100", "owner": "service-owner-10"}
		resp := server.GetServices(context.Background(), filters)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if len(resp.Services) != 1 {
			t.Fatalf("error: %d", len(resp.Services))
		}
	})
}

// 异常场景
func TestGetServices4(t *testing.T) {
	t.Run("查询服务列表，新建一批服务，删除部分，再查询，可以过滤掉删除的", func(t *testing.T) {
		total := 50
		for i := 0; i < total; i++ {
			serviceReq, serviceResp := createCommonService(t, i+5)
			defer cleanServiceName(serviceReq.GetName().GetValue(), serviceReq.GetNamespace().GetValue())
			if i%2 == 0 {
				removeCommonServices(t, []*api.Service{serviceResp})
			}
		}

		query := map[string]string{
			"offset": "0",
			"limit":  "100",
			"name":   "test-service-*",
		}
		resp := server.GetServices(context.Background(), query)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.Info.GetValue())
		}
		if resp.GetSize().GetValue() == uint32(total/2) {
			t.Logf("pass")
		} else {
			t.Fatalf("error: %d", resp.GetSize().GetValue())
		}
	})
	// 新建几个服务，不同metadata
	t.Run("根据metadata可以过滤services", func(t *testing.T) {
		service1 := genMainService(1)
		service1.Metadata = map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		service2 := genMainService(2)
		service2.Metadata = map[string]string{
			"key2": "value2",
			"key3": "value3",
		}
		service3 := genMainService(3)
		service3.Metadata = map[string]string{"key3": "value3"}
		if resp := server.CreateServices(defaultCtx, []*api.Service{service1, service2, service3}); !respSuccess(resp) {
			t.Fatalf("error: %+v", resp)
		}
		defer cleanServiceName(service1.GetName().GetValue(), service1.GetNamespace().GetValue())
		defer cleanServiceName(service2.GetName().GetValue(), service2.GetNamespace().GetValue())
		defer cleanServiceName(service3.GetName().GetValue(), service3.GetNamespace().GetValue())

		resps := server.GetServices(context.Background(), map[string]string{"keys": "key3", "values": "value3"})
		if len(resps.GetServices()) != 3 && resps.GetAmount().GetValue() != 3 {
			t.Fatalf("error: %d", len(resps.GetServices()))
		}
		resps = server.GetServices(context.Background(), map[string]string{"keys": "key2", "values": "value2"})
		if len(resps.GetServices()) != 2 && resps.GetAmount().GetValue() != 2 {
			t.Fatalf("error: %d", len(resps.GetServices()))
		}
		resps = server.GetServices(context.Background(), map[string]string{"keys": "key1", "values": "value1"})
		if len(resps.GetServices()) != 1 && resps.GetAmount().GetValue() != 1 {
			t.Fatalf("error: %d", len(resps.GetServices()))
		}
		resps = server.GetServices(context.Background(), map[string]string{"keys": "key1", "values": "value2"})
		if len(resps.GetServices()) != 0 && resps.GetAmount().GetValue() != 0 {
			t.Fatalf("error: %d", len(resps.GetServices()))
		}
	})
}

// 联合查询场景
func TestGetServices5(t *testing.T) {
	getServiceCheck := func(resp *api.BatchQueryResponse, amount, size uint32) {
		t.Logf("gocheck resp: %v", resp)
		convey.So(respSuccess(resp), convey.ShouldEqual, true)
		convey.So(resp.GetAmount().GetValue(), convey.ShouldEqual, amount)
		convey.So(resp.GetSize().GetValue(), convey.ShouldEqual, size)
	}
	convey.Convey("支持host查询到服务", t, func() {
		_, serviceResp := createCommonService(t, 200)
		defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
		instanceReq, instanceResp := createCommonInstance(t, serviceResp, 100)
		defer cleanInstance(instanceResp.GetId().GetValue())
		instanceReq, instanceResp = createCommonInstance(t, serviceResp, 101)
		defer cleanInstance(instanceResp.GetId().GetValue())
		query := map[string]string{
			"owner": "service-owner-200",
			"host":  instanceReq.GetHost().GetValue(),
		}
		convey.Convey("check-1", func() { getServiceCheck(server.GetServices(context.Background(), query), 1, 1) })

		// 同host的实例，对应一个服务，那么返回值也是一个
		instanceReq.Port.Value = 999
		resp := server.CreateInstance(defaultCtx, instanceReq)
		convey.So(respSuccess(resp), convey.ShouldEqual, true)
		defer cleanInstance(resp.Instance.GetId().GetValue())
		convey.Convey("check-2", func() { getServiceCheck(server.GetServices(context.Background(), query), 1, 1) })
	})
	convey.Convey("支持host和port配合查询服务", t, func() {
		host1 := "127.0.0.1"
		port1 := uint32(8081)
		host2 := "127.0.0.2"
		port2 := uint32(8082)
		_, serviceResp1 := createCommonService(t, 200)
		defer cleanServiceName(serviceResp1.GetName().GetValue(), serviceResp1.GetNamespace().GetValue())
		_, instanceResp1 := addHostPortInstance(t, serviceResp1, host1, port1)
		defer cleanInstance(instanceResp1.GetId().GetValue())
		_, serviceResp2 := createCommonService(t, 300)
		defer cleanServiceName(serviceResp2.GetName().GetValue(), serviceResp2.GetNamespace().GetValue())
		_, instanceResp2 := addHostPortInstance(t, serviceResp2, host1, port2)
		defer cleanInstance(instanceResp2.GetId().GetValue())
		_, serviceResp3 := createCommonService(t, 400)
		defer cleanServiceName(serviceResp3.GetName().GetValue(), serviceResp3.GetNamespace().GetValue())
		_, instanceResp3 := addHostPortInstance(t, serviceResp3, host2, port1)
		defer cleanInstance(instanceResp3.GetId().GetValue())
		_, serviceResp4 := createCommonService(t, 500)
		defer cleanServiceName(serviceResp4.GetName().GetValue(), serviceResp4.GetNamespace().GetValue())
		_, instanceResp4 := addHostPortInstance(t, serviceResp4, host2, port2)
		defer cleanInstance(instanceResp4.GetId().GetValue())

		query := map[string]string{
			"host": host1,
			"port": strconv.Itoa(int(port1)),
		}
		convey.Convey("check-1-1", func() {
			getServiceCheck(
				server.GetServices(context.Background(), query), 1, 1)
		})
		query["host"] = host1 + "," + host2
		convey.Convey("check-2-1", func() {
			getServiceCheck(
				server.GetServices(context.Background(), query), 2, 2)
		})
		query["port"] = fmt.Sprintf("%d,%d", port1, port2)
		convey.Convey("check-2-2", func() {
			getServiceCheck(
				server.GetServices(context.Background(), query), 4, 4)
		})
	})
	convey.Convey("多个服务，对应同个host，返回多个服务", t, func() {
		count := 10
		var instance *api.Instance
		for i := 0; i < count; i++ {
			_, serviceResp := createCommonService(t, i)
			defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
			_, instanceResp := createCommonInstance(t, serviceResp, 100)
			defer cleanInstance(instanceResp.GetId().GetValue())
			instance = instanceResp
			_, instanceResp = createCommonInstance(t, serviceResp, 202)
			defer cleanInstance(instanceResp.GetId().GetValue())
		}
		query := map[string]string{
			"host":  instance.GetHost().GetValue(),
			"limit": "5",
		}
		convey.Convey("check-1", func() {
			getServiceCheck(
				server.GetServices(context.Background(), query), uint32(count), 5)
		})
	})
}

// 测试更新服务
func TestUpdateService(t *testing.T) {
	_, serviceResp := createCommonService(t, 200)
	defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("正常更新服务，所有属性都生效", func(t *testing.T) {
		updateReq := &api.Service{
			Name:      serviceResp.Name,
			Namespace: serviceResp.Namespace,
			Metadata: map[string]string{
				"new-key":   "1",
				"new-key-2": "2",
				"new-key-3": "3",
			},
			Ports:      utils.NewStringValue("new-ports"),
			Business:   utils.NewStringValue("new-business"),
			Department: utils.NewStringValue("new-business"),
			CmdbMod1:   utils.NewStringValue("new-cmdb-mod1"),
			CmdbMod2:   utils.NewStringValue("new-cmdb-mo2"),
			CmdbMod3:   utils.NewStringValue("new-cmdb-mod3"),
			Comment:    utils.NewStringValue("new-comment"),
			Owners:     utils.NewStringValue("new-owner"),
			Token:      serviceResp.Token,
		}
		resp := server.UpdateService(defaultCtx, updateReq)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}

		// get service
		query := map[string]string{
			"name":      updateReq.GetName().GetValue(),
			"namespace": updateReq.GetNamespace().GetValue(),
		}
		services := server.GetServices(context.Background(), query)
		if !respSuccess(services) {
			t.Fatalf("error: %s", services.GetInfo().GetValue())
		}
		if services.GetSize().GetValue() != 1 {
			t.Fatalf("error: %d", services.GetSize().GetValue())
		}

		serviceCheck(t, updateReq, services.GetServices()[0])
	})
	t.Run("更新服务，metadata数据个数太多，报错", func(t *testing.T) {
		serviceResp.Metadata = make(map[string]string)
		for i := 0; i < service.MaxMetadataLength+1; i++ {
			serviceResp.Metadata[fmt.Sprintf("update-%d", i)] = "abc"
		}
		if resp := server.UpdateService(defaultCtx, serviceResp); !respSuccess(resp) {
			t.Logf("pass: %s", resp.GetInfo().GetValue())
		} else {
			t.Fatalf("error")
		}
	})
	t.Run("更新服务，metadata为空，长度为0，则删除所有metadata", func(t *testing.T) {
		serviceResp.Metadata = make(map[string]string)
		if resp := server.UpdateService(defaultCtx, serviceResp); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		getResp := server.GetServices(context.Background(), map[string]string{"name": serviceResp.Name.Value})
		if !respSuccess(getResp) {
			t.Fatalf("error: %s", getResp.GetInfo().GetValue())
		}
		if len(getResp.Services[0].Metadata) != 0 {
			t.Fatalf("error: %d", len(getResp.Services[0].Metadata))
		}
	})
	t.Run("更新服务，不允许更新别名", func(t *testing.T) {
		aliasResp := createCommonAlias(serviceResp, "update.service.alias.xxx", defaultAliasNs, api.AliasType_DEFAULT)
		defer cleanServiceName(aliasResp.Alias.Alias.Value, serviceResp.Namespace.Value)

		aliasService := &api.Service{
			Name:       aliasResp.Alias.Alias,
			Namespace:  serviceResp.Namespace,
			Department: utils.NewStringValue("123"),
			Token:      serviceResp.Token,
		}
		if resp := server.UpdateService(defaultCtx, aliasService); respSuccess(resp) {
			t.Fatalf("error: update alias success")
		} else {
			t.Logf("update alias return: %s", resp.GetInfo().GetValue())
		}
	})
}

// 服务更新，noChange测试
func TestNoNeedUpdateService(t *testing.T) {
	_, serviceResp := createCommonService(t, 500)
	defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("数据没有任意变更，返回不需要变更", func(t *testing.T) {
		resp := server.UpdateService(defaultCtx, serviceResp)
		if resp.GetCode().GetValue() != api.NoNeedUpdate {
			t.Fatalf("error: %+v", resp)
		}
	})
	req := &api.Service{
		Name:      serviceResp.Name,
		Namespace: serviceResp.Namespace,
		Token:     serviceResp.Token,
	}
	t.Run("metadata为空，不需要变更", func(t *testing.T) {
		req.Metadata = nil
		if resp := server.UpdateService(defaultCtx, req); resp.GetCode().GetValue() != api.NoNeedUpdate {
			t.Fatalf("error: %+v", resp)
		}
		req.Comment = serviceResp.Comment
		if resp := server.UpdateService(defaultCtx, req); resp.GetCode().GetValue() != api.NoNeedUpdate {
			t.Fatalf("error: %+v", resp)
		}
	})
	t.Run("metadata不为空，但是没变更，也不需要更新", func(t *testing.T) {
		req.Metadata = serviceResp.Metadata
		if resp := server.UpdateService(defaultCtx, req); resp.GetCode().GetValue() != api.NoNeedUpdate {
			t.Fatalf("error: %+v", resp)
		}
	})
	t.Run("其他字段更新，metadata没有更新，不需要更新metadata", func(t *testing.T) {
		req.Metadata = serviceResp.Metadata
		req.Comment = utils.NewStringValue("1357986420")
		if resp := server.UpdateService(defaultCtx, req); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}
	})
	t.Run("只有一个字段变更，service就执行变更操作", func(t *testing.T) {
		baseReq := api.Service{
			Name:      serviceResp.Name,
			Namespace: serviceResp.Namespace,
			Token:     serviceResp.Token,
		}

		r := baseReq
		r.Ports = utils.NewStringValue("90909090")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.Business = utils.NewStringValue("new-business")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.Department = utils.NewStringValue("new-department-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.CmdbMod1 = utils.NewStringValue("new-CmdbMod1-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.CmdbMod2 = utils.NewStringValue("new-CmdbMod2-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.CmdbMod3 = utils.NewStringValue("new-CmdbMod3-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.Comment = utils.NewStringValue("new-Comment-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}

		r = baseReq
		r.Owners = utils.NewStringValue("new-Owners-1")
		if resp := server.UpdateService(defaultCtx, &r); resp.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("error: %+v", resp)
		}
	})
}

// 测试serviceToken相关的操作
func TestServiceToken(t *testing.T) {
	_, serviceResp := createCommonService(t, 200)
	defer cleanServiceName(serviceResp.GetName().GetValue(), serviceResp.GetNamespace().GetValue())
	t.Run("可以正常获取serviceToken", func(t *testing.T) {
		req := &api.Service{
			Name:      serviceResp.GetName(),
			Namespace: serviceResp.GetNamespace(),
			Token:     serviceResp.GetToken(),
		}

		resp := server.GetServiceToken(defaultCtx, req)
		if !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		}
		if resp.GetService().GetToken().GetValue() != serviceResp.GetToken().GetValue() {
			t.Fatalf("error")
		}
	})

	t.Run("获取别名的token，返回源服务的token", func(t *testing.T) {
		aliasResp := createCommonAlias(serviceResp, "get.token.xxx", defaultAliasNs, api.AliasType_DEFAULT)
		defer cleanServiceName(aliasResp.Alias.Alias.Value, serviceResp.Namespace.Value)
		t.Logf("%+v", aliasResp)

		req := &api.Service{
			Name:      aliasResp.Alias.Alias,
			Namespace: serviceResp.GetNamespace(),
			Token:     serviceResp.GetToken(),
		}
		t.Logf("%+v", req)
		if resp := server.GetServiceToken(defaultCtx, req); !respSuccess(resp) {
			t.Fatalf("error: %s", resp.GetInfo().GetValue())
		} else if resp.GetService().GetToken().GetValue() != serviceResp.GetToken().GetValue() {
			t.Fatalf("error")
		}
	})

	t.Run("可以正常更新serviceToken", func(t *testing.T) {
		resp := server.UpdateServiceToken(defaultCtx, serviceResp)
		if !respSuccess(resp) {
			t.Fatalf("error :%s", resp.GetInfo().GetValue())
		}
		if resp.GetService().GetToken().GetValue() == serviceResp.GetToken().GetValue() {
			t.Fatalf("error: %s %s", resp.GetService().GetToken().GetValue(),
				serviceResp.GetToken().GetValue())
		}
		serviceResp.Token.Value = resp.Service.Token.Value // set token
	})

	t.Run("alias不允许更新token", func(t *testing.T) {
		aliasResp := createCommonAlias(serviceResp, "update.token.xxx", defaultAliasNs, api.AliasType_DEFAULT)
		defer cleanServiceName(aliasResp.Alias.Alias.Value, serviceResp.Namespace.Value)

		req := &api.Service{
			Name:      aliasResp.Alias.Alias,
			Namespace: serviceResp.Namespace,
			Token:     serviceResp.Token,
		}
		if resp := server.UpdateServiceToken(defaultCtx, req); respSuccess(resp) {
			t.Fatalf("error")
		}
	})
}

// 测试response格式化
func TestFormatBatchWriteResponse(t *testing.T) {
	t.Run("同样的错误码，返回一个错误码4XX", func(t *testing.T) {
		responses := api.NewBatchWriteResponse(api.ExecuteSuccess)
		for i := 0; i < 10; i++ {
			responses.Collect(api.NewResponse(api.NotFoundService))
		}

		responses = api.FormatBatchWriteResponse(responses)
		if responses.GetCode().GetValue() != api.NotFoundService {
			t.Fatalf("%+v", responses)
		}
	})
	t.Run("同样的错误码，返回一个错误码5XX", func(t *testing.T) {
		responses := api.NewBatchWriteResponse(api.ExecuteSuccess)
		for i := 0; i < 10; i++ {
			responses.Collect(api.NewResponse(api.StoreLayerException))
		}

		responses = api.FormatBatchWriteResponse(responses)
		if responses.GetCode().GetValue() != api.StoreLayerException {
			t.Fatalf("%+v", responses)
		}
	})
	t.Run("有5XX和2XX，返回5XX", func(t *testing.T) {
		responses := api.NewBatchWriteResponse(api.ExecuteSuccess)
		responses.Collect(api.NewResponse(api.ExecuteSuccess))
		responses.Collect(api.NewResponse(api.NotFoundNamespace))
		responses.Collect(api.NewResponse(api.ParseRateLimitException))
		responses.Collect(api.NewResponse(api.ParseException))
		responses = api.FormatBatchWriteResponse(responses)
		if responses.GetCode().GetValue() != api.ExecuteException {
			t.Fatalf("%+v", responses)
		}
	})
	t.Run("没有5XX，有4XX，返回4XX", func(t *testing.T) {
		responses := api.NewBatchWriteResponse(api.ExecuteSuccess)
		responses.Collect(api.NewResponse(api.ExecuteSuccess))
		responses.Collect(api.NewResponse(api.NotFoundNamespace))
		responses.Collect(api.NewResponse(api.NoNeedUpdate))
		responses.Collect(api.NewResponse(api.InvalidInstanceID))
		responses.Collect(api.NewResponse(api.ExecuteSuccess))
		responses = api.FormatBatchWriteResponse(responses)
		if responses.GetCode().GetValue() != api.BadRequest {
			t.Fatalf("%+v", responses)
		}
	})
	t.Run("全是2XX", func(t *testing.T) {
		responses := api.NewBatchWriteResponse(api.ExecuteSuccess)
		responses.Collect(api.NewResponse(api.ExecuteSuccess))
		responses.Collect(api.NewResponse(api.NoNeedUpdate))
		responses.Collect(api.NewResponse(api.DataNoChange))
		responses.Collect(api.NewResponse(api.NoNeedUpdate))
		responses.Collect(api.NewResponse(api.ExecuteSuccess))
		responses = api.FormatBatchWriteResponse(responses)
		if responses.GetCode().GetValue() != api.ExecuteSuccess {
			t.Fatalf("%+v", responses)
		}
	})
}

// test对service字段进行校验
func TestCheckServiceFieldLen(t *testing.T) {
	service := genMainService(400)
	t.Run("服务名超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldName := service.Name
		service.Name = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Name = oldName
		if resp.Code.Value != api.InvalidServiceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("命名空间超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldNameSpace := service.Namespace
		service.Namespace = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Namespace = oldNameSpace
		if resp.Code.Value != api.InvalidNamespaceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("Metadata超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldMetadata := service.Metadata
		oldMetadata[str] = str
		resp := server.CreateService(defaultCtx, service)
		service.Metadata = make(map[string]string)
		if resp.Code.Value != api.InvalidMetadata {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务ports超长", func(t *testing.T) {
		str := genSpecialStr(8193)
		oldPort := service.Ports
		service.Ports = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Ports = oldPort
		if resp.Code.Value != api.InvalidServicePorts {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务Business超长", func(t *testing.T) {
		str := genSpecialStr(129)
		oldBusiness := service.Business
		service.Business = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Business = oldBusiness
		if resp.Code.Value != api.InvalidServiceBusiness {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务-部门超长", func(t *testing.T) {
		str := genSpecialStr(1025)
		oldDepartment := service.Department
		service.Department = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Department = oldDepartment
		if resp.Code.Value != api.InvalidServiceDepartment {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务cmdb超长", func(t *testing.T) {
		str := genSpecialStr(1025)
		oldCMDB := service.CmdbMod1
		service.CmdbMod1 = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.CmdbMod1 = oldCMDB
		if resp.Code.Value != api.InvalidServiceCMDB {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务comment超长", func(t *testing.T) {
		str := genSpecialStr(1025)
		oldComment := service.Comment
		service.Comment = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Comment = oldComment
		if resp.Code.Value != api.InvalidServiceComment {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务owner超长", func(t *testing.T) {
		str := genSpecialStr(1025)
		oldOwner := service.Owners
		service.Owners = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Owners = oldOwner
		if resp.Code.Value != api.InvalidServiceOwners {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("服务token超长", func(t *testing.T) {
		str := genSpecialStr(2049)
		oldToken := service.Token
		service.Token = utils.NewStringValue(str)
		resp := server.CreateService(defaultCtx, service)
		service.Token = oldToken
		if resp.Code.Value != api.InvalidServiceToken {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("检测字段为空指针", func(t *testing.T) {
		oldName := service.Name
		service.Name = nil
		resp := server.CreateService(defaultCtx, service)
		service.Name = oldName
		if resp.Code.Value != api.InvalidServiceName {
			t.Fatalf("%+v", resp)
		}
	})
	t.Run("检测字段为空", func(t *testing.T) {
		oldName := service.Name
		service.Name = utils.NewStringValue("")
		resp := server.CreateService(defaultCtx, service)
		service.Name = oldName
		if resp.Code.Value != api.InvalidServiceName {
			t.Fatalf("%+v", resp)
		}
	})
}
