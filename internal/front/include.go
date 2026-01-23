// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package front

import (
	"fmt"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

var Include = gox.Elem(func(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	conf := core.Conf()
	registry := core.ImportRegistry()
	style := registry.MainStyle()
	script := registry.MainScript()
	if err := cur.InitVoid("link"); err != nil {
		return err
	}
	{
		if err := cur.AttrSet("rel", "stylesheet"); err != nil {
			return err
		}
		if err := cur.AttrSet("href", fmt.Sprintf("/%s.d00r.css", style.HashString())); err != nil {
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
		if err := cur.AttrSet("src", fmt.Sprintf("/%s.d00r.js", script.HashString())); err != nil {
			return err
		}
		if err := cur.AttrSet("data-id", core.InstanceID()); err != nil {
			return err
		}
		if err := cur.AttrSetAny("data-root", core.RootID()); err != nil {
			return err
		}
		if err := cur.AttrSetAny("data-ttl", conf.InstanceTTL.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSetAny("data-disconnect", conf.DisconnectHiddenTimer.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSetAny("data-request", conf.RequestTimeout.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSetAny("data-ping", conf.SolitairePing.Milliseconds()); err != nil {
			return err
		}
		if err := cur.AttrSetBool("data-detached", core.Detached()); err != nil {
			return err
		}
		lic := core.License()
		if lic != nil {
			licInfo := fmt.Sprintf("%s:%s:%s", lic.GetId(), lic.GetTier().String(), lic.GetDomain())
			if err := cur.AttrSetAny("data-license", licInfo); err != nil {
				return err
			}
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

