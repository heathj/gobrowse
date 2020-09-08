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

func (e *Event) ComposedPath() []EventTarget                           { return nil }
func (e *Event) StopPropagation()                                      {}
func (e *Event) StopImmediatePropagation()                             {}
func (e *Event) PreventDefault()                                       {}
func (e *Event) InitEvent(eventType webidl.DOMString, options ...bool) {}
