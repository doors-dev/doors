// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doors

import (
	"github.com/doors-dev/doors/internal/common"
)

// ErrPathModel reports an invalid path model definition: bad tags, wrong
// field types, or unexported fields. Returned by [NewLocation] and matched
// with [errors.Is].
var ErrPathModel = common.ErrPathModel

// ErrPathEncode reports a model that failed to encode into a location: no
// variant selected or an empty required capture. Returned by [NewLocation]
// and location actions, matched with [errors.Is].
var ErrPathEncode = common.ErrPathEncode

// ErrExecution reports an action that reached the browser and failed there.
// Delivered on completion channels of [Call] and door operations, matched
// with [errors.Is].
var ErrExecution = common.ErrExecution

// ErrTerminated reports an instance that ended before the operation could
// complete. Delivered on completion channels, matched with [errors.Is].
var ErrTerminated = common.ErrTerminated
