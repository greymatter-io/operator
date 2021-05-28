package controllers

type gmImages struct {
	Control    string
	ControlAPI string
	Proxy      string
}

var gmVersionMap = map[string]gmImages{
	"1.3": {
		Control:    "1.5.3",
		ControlAPI: "1.5.4",
		Proxy:      "1.5.1",
	},
	"1.2": {
		Control:    "1.4.2",
		ControlAPI: "1.4.1",
		Proxy:      "1.4.0",
	},
}
