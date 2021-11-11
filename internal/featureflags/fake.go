package featureflags

type Fake struct {
	Result bool
}


func (f *Fake) IsEnabled(flag string) bool {
	return f.Result	
}
