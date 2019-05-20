package announcers

// Fake is a fake announcer
type Fake struct {
	event *AvailableEvent
}

// Announce does nothing
func (f *Fake) Announce(e *AvailableEvent) {
	f.event = e
}