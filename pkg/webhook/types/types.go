package types

type Delivery struct {
	HookId                     string
	Event                      string
	Delivery                   string
	Signature                  string
	Signature256               string
	HookInstallationTargetType string
	HookInstallationTargetId   string
}
