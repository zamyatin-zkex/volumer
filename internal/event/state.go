package event

type StateSaved struct {
	Offset int64
}

type StateRestored struct {
	Offset int64
}
