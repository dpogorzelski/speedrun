package cloud

type Instance struct {
	PublicAddress  string
	PrivateAddress string
	Name           string
}

func (i Instance) GetAddress(private bool) string {
	if private {
		return i.PrivateAddress
	}

	return i.PublicAddress
}
