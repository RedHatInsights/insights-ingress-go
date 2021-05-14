package version

// IngressVersion is the json structure for the /version endpoint
type IngressVersion struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}
