package driver

// Enabler interface implemented to provide customized logging recorder filtering.
type Enabler interface {
	Enabled(*Recorder) bool
}

type denyAll struct{}
type acceptAll struct{}
type atAboveLevel struct{ atAbove int }
type matchLevel struct{ match int }
type rangeLevel struct{ min, max int }

// NewDenyAll return Enabler which drops all logging recorder.
func NewDenyAll() Enabler { return &denyAll{} }

// NewAcceptAll return Enabler which accepts all logging recorder.
func NewAcceptAll() Enabler { return &acceptAll{} }

// NewAtAboveLevel return a very simple Enabler
// which accepts logging recorder's level at or above the value of the n level.
func NewAtAboveLevel(n int) Enabler { return &atAboveLevel{n} }

// NewAtAboveLevel return a very simple Enabler
// which accepts logging recorder's level equals the value of the n level.
func NewMatachLevel(n int) Enabler { return &matchLevel{n} }

// NewRangeLevel return a very simple Enabler
// which accepts logging recorder's level match the range.
func NewRangeLevel(min, max int) Enabler { return &rangeLevel{} }

func (e *denyAll) Enabled(r *Recorder) bool      { return false }
func (e *acceptAll) Enabled(r *Recorder) bool    { return true }
func (e *atAboveLevel) Enabled(r *Recorder) bool { return (r.Level >= e.atAbove) }
func (e *matchLevel) Enabled(r *Recorder) bool   { return (r.Level == e.match) }
func (e *rangeLevel) Enabled(r *Recorder) bool   { return (r.Level >= e.min && r.Level <= e.max) }
