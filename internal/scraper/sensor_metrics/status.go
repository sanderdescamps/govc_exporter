package sensormetrics

type Status struct {
	failed       bool
	count        float64
	successCount float64
}

func (st *Status) Set(success bool) {
	if success {
		st.Success()
	} else {
		st.Fail()
	}
}

func (st *Status) Get() bool {
	return !st.failed
}

func (st *Status) GetFloat64() float64 {
	if st.failed {
		return 0.0
	}
	return 1.0
}

func (st *Status) Success() {
	st.count += 1
	st.successCount += 1
	st.failed = false
}

func (st *Status) Fail() {
	st.count += 1
	st.failed = true
}

func (st *Status) SuccessRate() float64 {
	return st.successCount / st.count
}
