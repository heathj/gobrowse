package spec

//https:domspec.whatwg.org/#eventtarget
type EventTarget struct {
}

func (et *EventTarget) AddEventListener()    {}
func (et *EventTarget) RemoveEventListener() {}
func (et *EventTarget) DispatchEvent()       {}
