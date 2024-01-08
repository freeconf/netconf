package netconf

// Implements request in RFC5277 for events from a "named" event
// stream.  This presumably was from before YANG driven data models where each
// event stream had a unique identifier known as the data path.

type NamedStreams struct {
	Predefined map[string]string
}
