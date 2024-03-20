package httphandlers

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
}
