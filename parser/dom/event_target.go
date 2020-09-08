package dom

//https://dom.spec.whatwg.org/#eventtarget
type EventTarget struct {
}

func (et *EventTarget) AddEventListener()    {}
func (et *EventTarget) RemoveEventListener() {}
func (et *EventTarget) DispatchEvent()       {}
