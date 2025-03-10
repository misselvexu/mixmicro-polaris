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

package timewheel

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

// test timewheel task run
func TestTaskRun1(t *testing.T) {
	tw := New(time.Second, 5, "test tw")
	tw.Start()
	callback := func(data interface{}) {
		fmt.Println(data.(string))
	}

	t.Logf("add task time:%d", time.Now().Unix())
	for i := 0; i < 10; i++ {
		tw.AddTask(1000, "polaris 1s "+strconv.Itoa(i), callback)
	}
	t.Logf("add task time end:%d", time.Now().Unix())

	time.Sleep(2 * time.Second)
	t.Logf("add task time:%d", time.Now().Unix())
	for i := 0; i < 10; i++ {
		tw.AddTask(3000, "polaris 3s "+strconv.Itoa(i), callback)
	}
	t.Logf("add task time end:%d", time.Now().Unix())

	time.Sleep(5 * time.Second)
	t.Logf("add task time:%d", time.Now().Unix())
	for i := 0; i < 10; i++ {
		tw.AddTask(10000, "polaris 10s "+strconv.Itoa(i), callback)
	}
	t.Logf("add task time end:%d", time.Now().Unix())
	time.Sleep(15 * time.Second)

	tw.Stop()
}

// test timewheel task run
func TestTaskRun2(t *testing.T) {
	tw := New(time.Second, 5, "test tw")
	tw.Start()
	callback := func(data interface{}) {
		now := time.Now().Unix()
		if now != 3123124121 {
			_ = fmt.Sprintf("%s%+v", data.(string), time.Now())
		} else {
			_ = fmt.Sprintf("%s%+v", data.(string), time.Now())
		}
	}

	t.Logf("add task time:%d", time.Now().Unix())
	for i := 0; i < 50000; i++ {
		tw.AddTask(3000, "polaris 3s "+strconv.Itoa(i), callback)
	}
	t.Logf("add task time end:%d", time.Now().Unix())
	time.Sleep(8)

	tw.Stop()
}

// test timewheel task run
func TestTaskRunBoth(t *testing.T) {
	tw := New(time.Second, 5, "test tw")
	tw.Start()
	callback := func(data interface{}) {
		fmt.Println(data.(string))
	}

	for i := 0; i < 10; i++ {
		go tw.AddTask(1000, "polaris 1s_"+strconv.Itoa(i), callback)
		go tw.AddTask(3000, "polaris 3s_"+strconv.Itoa(i), callback)
		go tw.AddTask(7000, "polaris 10s_"+strconv.Itoa(i), callback)
	}
	time.Sleep(12 * time.Second)
	tw.Stop()
}

// timewheel task struct
type Info struct {
	id  string
	ttl int
	ms  int64
}

// bench-test timewheel task add
func BenchmarkAddTask1(t *testing.B) {
	tw := New(time.Second, 5, "test tw")
	info := &Info{
		"abcdefghijklmnopqrstuvwxyz",
		2,
		time.Now().Unix(),
	}

	callback := func(data interface{}) {
		dataInfo := data.(*Info)
		if dataInfo.ms < time.Now().Unix() {
			fmt.Println("overtime")
		}
	}

	// t.N = 100000
	t.SetParallelism(10000)
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tw.AddTask(2000, info, callback)
		}
	})
}

// bench-test timewheel task add
// use 2 slot
func BenchmarkAddTask2(t *testing.B) {
	tw := New(time.Second, 5, "test tw")
	info := &Info{
		"abcdefghijklmnopqrstuvwxyz",
		2,
		time.Now().Unix(),
	}

	callback := func(data interface{}) {
		dataInfo := data.(*Info)
		if dataInfo.ms < time.Now().Unix() {
			fmt.Println("overtime")
		}
	}

	t.SetParallelism(10000)
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tw.AddTask(2000, info, callback)
			tw.AddTask(3000, info, callback)
		}
	})
}

// bench-test timewheel task add
// use 2 timewheel
func BenchmarkAddTask3(t *testing.B) {
	tw := New(time.Second, 5, "test tw")
	tw2 := New(time.Second, 5, "test tw")

	info := &Info{
		"abcdefghijklmnopqrstuvwxyz",
		2,
		time.Now().Unix(),
	}

	callback := func(data interface{}) {
		dataInfo := data.(*Info)
		if dataInfo.ms < time.Now().Unix() {
			fmt.Println("overtime")
		}
	}

	t.SetParallelism(10000)
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tw.AddTask(2000, info, callback)
			tw2.AddTask(2000, info, callback)
		}
	})
}

// result:select random get ch
func TestSelect(t *testing.T) {
	ch := make(chan int, 20)
	ch2 := make(chan int, 20)
	stopCh := make(chan bool)

	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			ch2 <- i + 20
		}
		time.Sleep(10 * time.Second)
		close(stopCh)
	}()

	for {
		select {
		case i := <-ch:
			fmt.Println(i)
			time.Sleep(time.Second)
		case i := <-ch2:
			fmt.Println(i)
			time.Sleep(time.Second)
		case <-stopCh:
			for i := range ch {
				fmt.Println(i)
			}
			return
		}
	}
}
