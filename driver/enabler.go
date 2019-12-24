package driver

// Enabler interface implemented to provide customized logging recorder filtering.
type Enabler interface {
	Enabled(*Recorder) bool
}

type denyAll struct{}
type acceptAll struct{}
type atAbove struct{ atAbove int }
type matchLevel struct{ match int }
type rangeLevel struct{ min, max int }

// DenyAll return Enabler which drops all logging recorder.
func DenyAll() Enabler { return &denyAll{} }

// AcceptAll return Enabler which accepts all logging recorder.
func AcceptAll() Enabler { return &acceptAll{} }

// AtAbove return a very simple Enabler
// which accepts logging recorder's level at or above the value of the n level.
func AtAbove(n int) Enabler { return &atAbove{n} }

// MatchLevel return a very simple Enabler
// which accepts logging recorder's level equals the value of the n level.
func MatchLevel(n int) Enabler { return &matchLevel{n} }

// RangeLevel return a very simple Enabler
// which accepts logging recorder's level match the range.
func RangeLevel(min, max int) Enabler { return &rangeLevel{} }

func (e *denyAll) Enabled(r *Recorder) bool      { return false }
func (e *acceptAll) Enabled(r *Recorder) bool    { return true }
func (e *atAbove) Enabled(r *Recorder) bool { return (r.Level >= e.atAbove) }
func (e *matchLevel) Enabled(r *Recorder) bool   { return (r.Level == e.match) }
func (e *rangeLevel) Enabled(r *Recorder) bool   { return (r.Level >= e.min && r.Level <= e.max) }
