package client

func deleteSoftValue(soft *bool) bool {
	return soft == nil || *soft
}
