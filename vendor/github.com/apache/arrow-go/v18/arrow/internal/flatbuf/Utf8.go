// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuf

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

// / Unicode with UTF-8 encoding
type Utf8 struct {
	_tab flatbuffers.Table
}

func GetRootAsUtf8(buf []byte, offset flatbuffers.UOffsetT) *Utf8 {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Utf8{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Utf8) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Utf8) Table() flatbuffers.Table {
	return rcv._tab
}

func Utf8Start(builder *flatbuffers.Builder) {
	builder.StartObject(0)
}
func Utf8End(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
