package pine

import (
	"time"

	"github.com/pkg/errors"
)

type Series interface {
	AddIndicator(name string, i Indicator) error
	AddExec(v TPQ) error
	AddOHLCV(v OHLCV) error
	GetValueForInterval(t time.Time) *Interval
}

type Interval struct {
	StartTime  time.Time
	OHLCV      *OHLCV
	Value      float64
	Indicators map[string]*float64
}

type series struct {
	items    map[string]Indicator
	lastExec TPQ
	lastOHLC *OHLCV
	opts     SeriesOpts
	values   []OHLCV
	timemap  map[time.Time]*OHLCV
}

// NewSeries generates new OHLCV serie
func NewSeries(ohlcv []OHLCV, opts SeriesOpts) (Series, error) {
	err := opts.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "error validating seriesopts")
	}
	tm := make(map[time.Time]*OHLCV)
	s := &series{
		items:   make(map[string]Indicator),
		opts:    opts,
		timemap: tm,
		values:  make([]OHLCV, 0, opts.Max),
	}
	s.initValues(ohlcv)
	return s, nil
}

func (s *series) initValues(values []OHLCV) {
	for _, v := range values {
		s.insertInterval(v)
	}
}

func (s *series) insertInterval(v OHLCV) {
	t := s.getLastIntervalFromTime(v.S)
	v.S = t
	_, ok := s.timemap[t]
	if !ok {
		s.values = append(s.values, v)
		s.timemap[t] = &v
		s.lastOHLC = &v
	}
}

func (s *series) updateIndicators(v OHLCV) error {
	for _, ind := range s.items {
		if err := ind.Update(v); err != nil {
			return errors.Wrap(err, "error updating indicator")
		}
	}
	return nil
}

func (s *series) getLastIntervalFromTime(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	st := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	m := s.getMultiplierDiff(t, st)
	return st.Add(time.Duration(m*s.opts.Interval) * time.Second)
}

func (s *series) getMultiplierDiff(t time.Time, st time.Time) int {
	diff := t.Sub(st).Seconds()
	return int(diff / float64(s.opts.Interval))
}

func (s *series) getOHLCV(t time.Time) *OHLCV {
	return s.timemap[t]
}
