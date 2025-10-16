package transip

import (
	"io"

	"github.com/pbergman/provider"
)

func (p *Provider) DebugOutputLevel() provider.OutputLevel {
	return p.DebugLevel
}

func (p *Provider) DebugOutput() io.Writer {
	return p.DebugOut
}

func (p *Provider) SetDebug(level provider.OutputLevel, writer io.Writer) {
	p.DebugLevel = level
	p.DebugOut = writer
}
