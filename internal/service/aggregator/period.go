package aggregator

import "time"

type Period time.Duration

type Periods map[string]time.Duration

func (p Periods) max() Period {
	if len(p) == 0 {
		return 0
	}

	var longest time.Duration
	for _, period := range p {
		if period > longest {
			longest = period
		}
	}
	return Period(longest)
}

func (p Period) buckets() int {
	return int(time.Duration(p).Seconds())
}
