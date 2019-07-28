package oracle

type event struct {
	canceled bool
	err      error
}

func (s *event) Cancel(err error) {
	s.err = err
	s.canceled = true
}
