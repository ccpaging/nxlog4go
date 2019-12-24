// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

func find(filters []*Filter, filter *Filter) int {
	for i, f := range filters {
		if f == filter {
			return i
		}
	}
	return -1
}

// Attach adds the filters to logger.
func (l *Logger) Attach(filters ...*Filter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if i := find(l.filters, f); i >= 0 {
			// Existed
			continue
		}
		l.filters = append(l.filters, f)
	}
}

// Detach removes the filters from logger.
func (l *Logger) Detach(filters ...*Filter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if i := find(l.filters, f); i >= 0 {
			// Existed
			l.filters = append(l.filters[:i], l.filters[i+1:]...)
		}
	}
}
