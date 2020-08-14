package parser

//https://dom.spec.whatwg.org/#eventtarget
type EventTarget struct {
}

func (et *EventTarget) addEventListener()    {}
func (et *EventTarget) removeEventListener() {}
func (et *EventTarget) dispatchEvent()       {}
