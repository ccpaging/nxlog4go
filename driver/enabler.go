package driver

type Enabler interface {
	Enabled(*Recorder) bool
}

type denyAll struct{}
type acceptAll struct{}
type atAboveLevel struct{ atAbove int }
type matchLevel struct{ match int }
type rangeLevel struct{ min, max int }

func NewDenyAll() Enabler                { return &denyAll{} }
func NewAcceptAll() Enabler              { return &acceptAll{} }
func NewAtAboveLevel(n int) Enabler      { return &atAboveLevel{n} }
func NewMatachLevel() Enabler            { return &matchLevel{} }
func NewRangeLevel(min, max int) Enabler { return &rangeLevel{} }

func (e *denyAll) Enabled(r *Recorder) bool      { return false }
func (e *acceptAll) Enabled(r *Recorder) bool    { return true }
func (e *atAboveLevel) Enabled(r *Recorder) bool { return (r.Level >= e.atAbove) }
func (e *matchLevel) Enabled(r *Recorder) bool   { return (r.Level == e.match) }
func (e *rangeLevel) Enabled(r *Recorder) bool   { return (r.Level >= e.min && r.Level <= e.max) }
