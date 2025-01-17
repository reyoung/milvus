// Copyright (C) 2019-2020 Zilliz. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under the License.

package paramtable

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var baseParams = BaseTable{}

func TestMain(m *testing.M) {
	baseParams.Init()
	code := m.Run()
	os.Exit(code)
}

//func TestMain

func TestGlobalParamsTable_SaveAndLoad(t *testing.T) {
	err1 := baseParams.Save("int", "10")
	assert.Nil(t, err1)

	err2 := baseParams.Save("string", "testSaveAndLoad")
	assert.Nil(t, err2)

	err3 := baseParams.Save("float", "1.234")
	assert.Nil(t, err3)

	r1, _ := baseParams.Load("int")
	assert.Equal(t, "10", r1)

	r2, _ := baseParams.Load("string")
	assert.Equal(t, "testSaveAndLoad", r2)

	r3, _ := baseParams.Load("float")
	assert.Equal(t, "1.234", r3)

	err4 := baseParams.Remove("int")
	assert.Nil(t, err4)

	err5 := baseParams.Remove("string")
	assert.Nil(t, err5)

	err6 := baseParams.Remove("float")
	assert.Nil(t, err6)
}

func TestGlobalParamsTable_LoadRange(t *testing.T) {
	_ = baseParams.Save("xxxaab", "10")
	_ = baseParams.Save("xxxfghz", "20")
	_ = baseParams.Save("xxxbcde", "1.1")
	_ = baseParams.Save("xxxabcd", "testSaveAndLoad")
	_ = baseParams.Save("xxxzhi", "12")

	keys, values, err := baseParams.LoadRange("xxxa", "xxxg", 10)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(keys))
	assert.Equal(t, "10", values[0])
	assert.Equal(t, "testSaveAndLoad", values[1])
	assert.Equal(t, "1.1", values[2])
	assert.Equal(t, "20", values[3])

	_ = baseParams.Remove("abc")
	_ = baseParams.Remove("fghz")
	_ = baseParams.Remove("bcde")
	_ = baseParams.Remove("abcd")
	_ = baseParams.Remove("zhi")
}

func TestGlobalParamsTable_Remove(t *testing.T) {
	err1 := baseParams.Save("RemoveInt", "10")
	assert.Nil(t, err1)

	err2 := baseParams.Save("RemoveString", "testRemove")
	assert.Nil(t, err2)

	err3 := baseParams.Save("RemoveFloat", "1.234")
	assert.Nil(t, err3)

	err4 := baseParams.Remove("RemoveInt")
	assert.Nil(t, err4)

	err5 := baseParams.Remove("RemoveString")
	assert.Nil(t, err5)

	err6 := baseParams.Remove("RemoveFloat")
	assert.Nil(t, err6)
}

func TestGlobalParamsTable_LoadYaml(t *testing.T) {
	err := baseParams.LoadYaml("milvus.yaml")
	assert.Nil(t, err)

	err = baseParams.LoadYaml("advanced/channel.yaml")
	assert.Nil(t, err)

	_, err = baseParams.Load("etcd.address")
	assert.Nil(t, err)
	_, err = baseParams.Load("pulsar.port")
	assert.Nil(t, err)
}

func TestBaseTable_ParseIntWithErr(t *testing.T) {
	var err error

	key1 := "ParseIntWithErrInt"
	err = baseParams.Save(key1, "10")
	assert.Nil(t, err)
	ten, err := baseParams.ParseIntWithErr(key1)
	assert.Nil(t, err)
	assert.Equal(t, 10, ten)

	key2 := "ParseIntWithErrInvalidInt"
	err = baseParams.Save(key2, "invalid")
	assert.Nil(t, err)
	_, err = baseParams.ParseIntWithErr(key2)
	assert.NotNil(t, err)
}
