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

package front

import (
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

var Include = gox.Elem(func(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	conf := core.Conf()
	registry := core.ResourceRegistry()
	pathMaker := core.PathMaker()
	if err := cur.InitVoid("link"); err != nil {
		return err
	}
	{
		if err := cur.AttrSet("rel", "stylesheet"); err != nil {
			return err
		}
		if err := cur.AttrSet("href", pathMaker.Resource(registry.MainStyle(), "d0r.css")); err != nil {
			return err
		}
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if err := cur.Init("script"); err != nil {
		return err
	}
	{
		if err := cur.AttrSet("src", pathMaker.Resource(registry.MainScript(), "d0r.js")); err != nil {
			return err
		}
		if err := cur.AttrSet("id", core.InstanceID()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-prefix", pathMaker.Prefix()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-root", core.RootID()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-ttl", conf.InstanceTTL.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-disconnect", conf.DisconnectHiddenTimer.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-request", conf.RequestTimeout.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSet("data-ping", conf.SolitairePing.Milliseconds()); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
	}
	if err := cur.Close(); err != nil {
		return err
	}
	return nil
})
