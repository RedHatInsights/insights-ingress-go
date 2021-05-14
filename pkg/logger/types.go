package logger

// CustomCloudwatch adds hostname and app name
type CustomCloudwatch struct {
	Hostname string
}

//Marshaler is an interface any type can implement to change its output in our production logs.
type Marshaler interface {
	MarshalLog() map[string]interface{}
}
