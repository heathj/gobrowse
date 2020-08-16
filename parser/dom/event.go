package dom

import "browser/parser/webidl"

type eventPhase uint

const (
	noneEventPhase eventPhase = iota
	capturingPhase
	atTargetPhase
	bubblingPhase
)

// https://dom.spec.whatwg.org/#interface-event
type Event struct {
	eventType        webidl.DOMString
	target           EventTarget
	srcElement       EventTarget
	currentTarget    EventTarget
	eventPhase       eventPhase
	cancelBubble     bool
	bubbles          bool
	cancelable       bool
	returnValue      bool
	defaultPrevented bool
	composed         bool
	isTrusted        bool
	timeStamp        webidl.DOMHighResTimeStamp
}

func (e *Event) composedPath() []EventTarget                           { return nil }
func (e *Event) stopPropagation()                                      {}
func (e *Event) stopImmediatePropagation()                             {}
func (e *Event) preventDefault()                                       {}
func (e *Event) initEvent(eventType webidl.DOMString, options ...bool) {}
