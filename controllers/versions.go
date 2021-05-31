package controllers

type gmImages struct {
	Control     string
	ControlAPI  string
	Proxy       string
	Catalog     string
	JwtSecurity string
}

var gmVersionMap = map[string]gmImages{
	"1.3": {
		Control:     "1.5.3",
		ControlAPI:  "1.5.4",
		Proxy:       "1.5.1",
		Catalog:     "1.2.2",
		JwtSecurity: "1.2.0",
	},
	"1.2": {
		Control:     "1.4.2",
		ControlAPI:  "1.4.1",
		Proxy:       "1.4.0",
		Catalog:     "1.0.7",
		JwtSecurity: "1.1.1",
	},
}
