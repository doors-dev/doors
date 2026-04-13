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

package door

import (
	"context"
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door/pipe"
	"github.com/doors-dev/doors/internal/shredder"
)

type node struct {
	ctx       context.Context
	initFrame shredder.ValveFrame
	door      *Door
	entity    any
}

func (n *node) Context() context.Context {
	return n.ctx
}

func (n *node) Output(io.Writer) error {
	return errors.New("door: used outside render pipeline")
}

func (n *node) contextParent() *tracker {
	return n.ctx.Value(ctex.KeyCore).(core.Core).Door().(*tracker)
}

func (n *node) init(prev *node) {
	switch ent := n.entity.(type) {
	case *replaceNode:
		n.initReplace(ent, prev)
	case *proxyNode:
		n.initProxy(ent, prev)
	case *editorNode:
		n.initEditor(prev)
	case *unmountNode:
		n.initUnmount(ent, prev)
	case *updateNode:
		n.initUpdate(ent, prev)
	case *redrawNode:
		n.initRedraw(ent, prev)
	case *rebaseNode:
		n.initRebase(ent, prev)
	}
}

func (n *node) initReplace(nr *replaceNode, prev *node) {
	prevEnt, ok := prev.entity.(nodeMount)
	if !ok {
		nr.Accept()
		return
	}

	nr.replaceId = prevEnt.ReplaceId()

	prev.killUnmount()

	n.writeReplace(
		nr,
		prevEnt.Tracker().parent,
		prevEnt.WriteFrame(),
	)
}

func (n *node) initProxy(np *proxyNode, prev *node) {

	prev.killRemove()

	np.tracker = newTracker(n.contextParent())
	np.contents = &contents{
		initializeFrame: &shredder.ValveFrame{},
	}
	np.initMountFrame()
	np.setReplaceId()
	switch ent := prev.entity.(type) {
	case *replaceNode:
		np.prevContents = nil
	case contentsNode:
		np.prevContents = ent.Contents()
	default:
		panic("unexpected node entity")
	}

	np.tracker.parent.addChild(n)
}

func (n *node) initUnmount(num *unmountNode, prev *node) {
	prev.killUnmount()

	switch ent := prev.entity.(type) {
	case *replaceNode:
		num.contents = &contents{
			initializeFrame: &shredder.ValveFrame{},
		}
		num.contents.initializeFrame.Activate()
	case *proxyNode:
		num.contents = ent.contents
		num.proxyElem = ent.elem
	case contentsNode:
		num.contents = ent.Contents()
	}
	nm, ok := prev.entity.(nodeMount)
	if !ok || !num.remove {
		num.Accept()
	}
	sendFrame := shredder.Join(true, nm.WriteFrame(), num.TaskFrame())
	defer sendFrame.Release()
	sendFrame.Run(nm.Tracker().parentContext(), nm.Tracker().runtime(), func(b bool) {
		if !b {
			num.Cancel()
			return
		}
		nm.Tracker().root.inst.Call(&call{
			ctx:     nm.Tracker().parentContext(),
			task:    num.taskNode,
			kind:    callReplace,
			id:      nm.Tracker().id,
			payload: pipe.EmptyPayload(),
		})
	})
}

func (n *node) initEditor(prev *node) {
	prev.killRemove()

	switch prevEnt := prev.entity.(type) {
	case *replaceNode:
		n.entity = prevEnt
	case *proxyNode:
		np := &proxyNode{
			mountNode: mountNode{
				tracker: newTracker(n.contextParent()),
				contents: &contents{
					initializeFrame: &shredder.ValveFrame{},
				},
			},
			elem:         prevEnt.elem,
			prevContents: prevEnt.contents,
		}
		np.initMountFrame()
		np.setReplaceId()
		n.entity = np

		np.tracker.parent.addChild(n)
	case *unmountNode:
		if prevEnt.wasProxy() {
			np := &proxyNode{
				mountNode: mountNode{
					tracker: newTracker(n.contextParent()),
					contents: &contents{
						initializeFrame: &shredder.ValveFrame{},
					},
				},
				elem:         prevEnt.proxyElem,
				prevContents: prevEnt.contents,
			}
			np.initMountFrame()
			np.setReplaceId()
			n.entity = np

			np.tracker.parent.addChild(n)
		} else {
			nu := &updateNode{
				mountNode: mountNode{
					tracker: newTracker(n.contextParent()),
					contents: &contents{
						initializeFrame: prevEnt.contents.initializeFrame,
						content:         prevEnt.contents.content,
						container:       prevEnt.contents.container,
					},
				},
			}
			nu.initMountFrame()
			n.entity = nu

			nu.tracker.parent.addChild(n)
		}
	case *updateNode:
		nu := &updateNode{
			mountNode: mountNode{
				tracker: newTracker(n.contextParent()),
				contents: &contents{
					initializeFrame: prevEnt.contents.initializeFrame,
					content:         prevEnt.contents.content,
					container:       prevEnt.contents.container,
				},
			},
		}
		nu.initMountFrame()
		n.entity = nu

		nu.tracker.parent.addChild(n)
	}
}

func (n *node) initRedraw(nr *redrawNode, prev *node) {
	switch prevEnt := prev.entity.(type) {
	case *replaceNode:
		n.entity = prevEnt
		nr.Report(errors.New("replaced door can't be reloaded"))
	case *unmountNode:
		n.entity = prevEnt
		nr.Report(errors.New("unmounted door can't be reloaded"))
	case *proxyNode:
		prev.killUnmount()

		np := &proxyNode{
			taskNode: nr.taskNode,
			mountNode: mountNode{
				tracker: newTracker(prevEnt.tracker.parent),
				contents: &contents{
					initializeFrame: &shredder.ValveFrame{},
				},
				mountFrame: prevEnt.mountFrame,
			},
			elem:         prevEnt.elem,
			prevContents: prevEnt.contents,
		}
		np.inheritReplaceId(prevEnt)
		n.entity = np

		np.tracker.parent.addChild(n)

		n.writeProxyReplace(np)
	case *updateNode:
		nu := &updateNode{
			taskNode: nr.taskNode,
			content:  prevEnt.contents.content,
		}
		n.entity = nu
		n.initUpdate(nu, prev)
	}
}

func (n *node) initRebase(nr *rebaseNode, prev *node) {
	switch prevEnt := prev.entity.(type) {
	case nodeMount:
		prev.killUnmount()
		np := &proxyNode{
			taskNode: nr.taskNode,
			mountNode: mountNode{
				tracker: newTracker(prevEnt.Tracker().parent),
				contents: &contents{
					initializeFrame: &shredder.ValveFrame{},
				},
				mountFrame: prevEnt.MountFrame(),
			},
			elem: nr.elem,
		}
		np.inheritReplaceId(prevEnt)
		n.entity = np

		np.tracker.parent.addChild(n)

		n.writeProxyReplace(np)
	default:
		nr.Accept()
		n.entity = &unmountNode{
			proxyElem: nr.elem,
		}
	}

}

func (n *node) initUpdate(nu *updateNode, prev *node) {
	switch prevEnt := prev.entity.(type) {
	case *replaceNode:
		num := &unmountNode{
			contents: &contents{
				container:       &Container{},
				initializeFrame: &shredder.ValveFrame{},
				content:         nu.content,
			},
		}
		num.contents.initializeFrame.Activate()
		n.entity = num
		nu.Accept()
	case *unmountNode:
		num := &unmountNode{
			contents: &contents{
				initializeFrame: prevEnt.Contents().initializeFrame,
				container:       prevEnt.Contents().container,
				content:         nu.content,
			},
		}
		n.entity = num
		nu.Accept()
	case nodeMount:
		nu.tracker = newTrackerFrom(prevEnt.Tracker())

		nu.contents = &contents{
			initializeFrame: prevEnt.Contents().initializeFrame,
			container:       prevEnt.Contents().container,
			content:         nu.content,
		}
		nu.mountFrame = prevEnt.MountFrame()

		prev.killReplace(n)

		n.writeUpdate(nu)
	}
}

func (n *node) writeReplace(nr *replaceNode, parentTracker *tracker, prevWriteFrame *shredder.ValveFrame) {
	thread := shredder.Thread{}
	renderFrame := shredder.Join(true, thread.Frame(), parentTracker.newRenderFrame())
	defer renderFrame.Release()
	sendFrame := shredder.Join(true, thread.Frame(), prevWriteFrame, nr.TaskFrame())
	defer sendFrame.Release()
	mountFrame := &shredder.ValveFrame{}
	pip := pipe.NewPipe(parentTracker.Context(), parentTracker.runtime(), renderFrame, mountFrame)
	renderFrame.Submit(parentTracker.ctx, parentTracker.runtime(), func(b bool) {
		if !b {
			return
		}
		pip.RenderContent(nr.content)
	})
	sendFrame.Run(parentTracker.ctx, parentTracker.runtime(), func(b bool) {
		defer mountFrame.Activate()
		if !b {
			nr.Cancel()
			return
		}
		payload := pip.Payload(parentTracker.inst().Conf().SolitaireDisableGzip)
		if e, ok := payload.(error); ok {
			nr.Report(e)
		}
		parentTracker.inst().Call(&call{
			ctx:     parentTracker.ctx,
			kind:    callReplace,
			id:      nr.replaceId,
			task:    nr.taskNode,
			payload: payload,
		})
	})
}

func (n *node) writeUpdate(nu *updateNode) {

	// 	nu.tracker.root.debug("UPDATE ", nu.tracker.id, nu.tracker.parent.id)
	thread := shredder.Thread{}
	renderFrame := shredder.Join(true, thread.Frame(), nu.tracker.newRenderFrame())
	defer renderFrame.Release()
	sendFrame := shredder.Join(true, thread.Frame(), nu.WriteFrame(), nu.TaskFrame())
	defer sendFrame.Release()
	mountFrame := &shredder.ValveFrame{}
	pip := pipe.NewPipe(nu.tracker.ctx, nu.tracker.runtime(), renderFrame, mountFrame)
	renderFrame.Submit(nu.tracker.ctx, nu.tracker.runtime(), func(b bool) {
		if !b {
			return
		}
		pip.RenderContent(nu.contents.content)
	})
	sendFrame.Run(nu.tracker.ctx, nu.tracker.runtime(), func(b bool) {
		defer mountFrame.Activate()
		if !b {
			nu.Cancel()
			return
		}
		payload := pip.Payload(nu.tracker.inst().Conf().SolitaireDisableGzip)
		if e, ok := payload.(error); ok {
			nu.Report(e)
			n.killUnmount()
		}
		nu.tracker.inst().Call(&call{
			ctx:     nu.tracker.ctx,
			kind:    callUpdate,
			id:      nu.tracker.id,
			task:    nu.taskNode,
			payload: payload,
		})
	})

}

func (n *node) writeProxyReplace(pn *proxyNode) {
	// pn.tracker.root.debug("PROXY_REPLACE ", pn.tracker.id, "-", pn.replaceId.Load(), pn.tracker.parent.id)
	thread := shredder.Thread{}
	renderFrame := shredder.Join(true, thread.Frame(), pn.tracker.newRenderFrame())
	defer renderFrame.Release()
	sendFrame := shredder.Join(true, thread.Frame(), pn.WriteFrame(), pn.TaskFrame())
	defer sendFrame.Release()
	mountFrame := &shredder.ValveFrame{}
	pip := pipe.NewPipe(pn.tracker.Context(), pn.tracker.runtime(), renderFrame, mountFrame)
	renderFrame.Submit(pn.tracker.parentContext(), pn.tracker.runtime(), func(b bool) {
		if !b {
			return
		}
		cont, ok := pip.RenderProxy(pn.elem)
		pn.contents.container = &cont
		if !pip.IsEmpty() || pn.prevContents == nil || !ok {
			pn.contents.container.Apply(pip, pn.tracker.containerContext(), pn.tracker.id, pn.tracker.parent.id)
			pn.contents.initializeFrame.Activate()
			return
		}
		prevReady := shredder.Join(false, renderFrame, pn.prevContents.initializeFrame)
		defer prevReady.Release()
		prevReady.Run(pn.tracker.Context(), pn.tracker.runtime(), func(b bool) {
			if pn.prevContents.content != nil {
				pn.contents.content = pn.prevContents.content
				pip.RenderContent(pn.contents.content)
			}
			pn.prevContents = nil
			pn.contents.container.Apply(pip, pn.tracker.containerContext(), pn.tracker.id, pn.tracker.parent.id)
			pn.contents.initializeFrame.Activate()
		})
	})
	sendFrame.Run(pn.tracker.ctx, pn.tracker.runtime(), func(b bool) {
		defer mountFrame.Activate()
		if !b {
			pn.Cancel()
			return
		}
		replaceId := pn.setReplaceId()
		if replaceId == 0 {
			pn.Cancel()
			return
		}
		payload := pip.Payload(pn.tracker.inst().Conf().SolitaireDisableGzip)
		if e, ok := payload.(error); ok {
			pn.Report(e)
			n.killUnmount()
		}
		pn.tracker.inst().Call(&call{
			ctx:     pn.tracker.parentContext(),
			kind:    callReplace,
			id:      replaceId,
			task:    pn.taskNode,
			payload: payload,
		})
	})
}

func (n *node) Render(pip pipe.Pipe) {
	branch := pip.Branch()
	n.initFrame.Run(pip.Context(), pip.Runtime(), func(b bool) {
		switch ent := n.entity.(type) {
		case *replaceNode:
			n.renderReplace(ent, pip, branch)
		case *updateNode:
			n.renderUpdate(ent, pip, branch)
		case *proxyNode:
			n.renderProxy(ent, pip, branch)
		default:
			panic("unexpected node entity to render")
		}
	})
}

func (n *node) renderReplace(nr *replaceNode, pip pipe.Pipe, branch pipe.Branch) {
	pip = pipe.NewPipe(pip.Context(), pip.Runtime(), pip.RenderFrame(), pip.FinalFrame())
	pip.RenderFrame().Run(pip.Context(), pip.Runtime(), func(b bool) {
		if !b {
			return
		}
		defer pip.Submit(branch)
		pip.RenderContent(nr.content)
	})

}

func (n *node) renderUpdate(nr *updateNode, pip pipe.Pipe, branch pipe.Branch) {
	nr.SetMountFrame(pip.FinalFrame())
	renderFrame := shredder.Join(true, pip.RenderFrame(), nr.Tracker().newRenderFrame(), nr.contents.initializeFrame)
	defer renderFrame.Release()
	pip = pipe.NewPipe(nr.tracker.Context(), nr.tracker.runtime(), renderFrame, pip.FinalFrame())
	renderFrame.Submit(nr.tracker.parentContext(), nr.tracker.runtime(), func(b bool) {
		if !b {
			return
		}
		defer func() {
			if err := pip.Submit(branch); err != nil {
				n.killUnmount()
			}
		}()
		pip.RenderContent(nr.contents.content)
		nr.contents.container.Apply(pip, nr.tracker.containerContext(), nr.tracker.id, nr.tracker.parent.id)
	})
}

func (n *node) renderProxy(pn *proxyNode, pip pipe.Pipe, branch pipe.Branch) {
	pn.SetMountFrame(pip.FinalFrame())
	renderFrame := shredder.Join(true, pip.RenderFrame(), pn.Tracker().newRenderFrame())
	defer renderFrame.Release()
	pip = pipe.NewPipe(pn.tracker.Context(), pn.tracker.runtime(), renderFrame, pip.FinalFrame())
	renderFrame.Submit(pn.tracker.parentContext(), pn.tracker.runtime(), func(b bool) {
		submit := false
		defer func() {
			if !submit {
				return
			}
			if err := pip.Submit(branch); err != nil {
				n.killUnmount()
			}
		}()
		cont, ok := pip.RenderProxy(pn.elem)
		pn.contents.container = &cont
		if !pip.IsEmpty() || pn.prevContents == nil || !ok {
			pn.prevContents = nil
			pn.contents.initializeFrame.Activate()
			pn.contents.container.Apply(pip, pn.tracker.containerContext(), pn.tracker.id, pn.tracker.parent.id)
			submit = true
			return
		}
		prevReady := shredder.Join(false, renderFrame, pn.prevContents.initializeFrame)
		defer prevReady.Release()
		prevReady.Run(pn.tracker.Context(), pn.tracker.runtime(), func(b bool) {
			defer func() {
				if err := pip.Submit(branch); err != nil {
					n.killUnmount()
				}
			}()
			if pn.prevContents.content != nil {
				pn.contents.content = pn.prevContents.content
				pip.RenderContent(pn.contents.content)
			}
			pn.prevContents = nil
			pn.contents.initializeFrame.Activate()
			pn.contents.container.Apply(pip, pn.tracker.containerContext(), pn.tracker.id, pn.tracker.parent.id)
		})
	})
}

func (n *node) killCascade() {
	nm, ok := n.entity.(nodeMount)
	if !ok {
		panic("must be mounted for cascade kill")
	}
	if !nm.Kill() {
		return
	}
	n.door.takeoverSelf(n, &node{
		ctx:  n.ctx,
		door: n.door,
		entity: &unmountNode{
			remove: false,
		},
	})
}

func (n *node) killReplace(new *node) {
	nm, ok := n.entity.(nodeMount)
	if !ok {
		panic("must be mounted to replace")
	}
	nm.Kill()
	nm.Tracker().parent.replaceChild(n, new)
}

func (n *node) killRemove() {
	nm, ok := n.entity.(nodeMount)
	if !ok {
		return
	}
	if nm.Kill() {
		nm.Tracker().parent.removeChild(n, nm.Tracker().id)
	}
	nm.WriteFrame().Run(nm.Tracker().parentContext(), nm.Tracker().runtime(), func(ok bool) {
		if !ok {
			return
		}
		nm.Tracker().root.inst.Call(&call{
			ctx:     nm.Tracker().parentContext(),
			kind:    callReplace,
			id:      nm.Tracker().id,
			payload: pipe.EmptyPayload(),
		})
	})
}

func (n *node) killUnmount() {
	nm, ok := n.entity.(nodeMount)
	if !ok {
		return
	}
	if !nm.Kill() {
		return
	}
	nm.Tracker().parent.removeChild(n, nm.Tracker().id)
}
