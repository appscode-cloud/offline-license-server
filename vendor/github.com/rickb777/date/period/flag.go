package period

func (period *Period) Set(p string) error {
	p2, err := Parse(p)
	if err != nil {
		return err
	}
	*period = p2
	return nil
}

func (period Period) Type() string {
	return "period"
}
