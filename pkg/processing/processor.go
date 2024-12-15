package processing

func NewProcessor() Processor {
	// XXX returns the simple log processor for now.
	// XXX this should be controlled through the ProcessorConfig
	// XXX as a follow up.
	return &logger{}
}
