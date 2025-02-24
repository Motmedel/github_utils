package types

const (
	HeaderGitHubHookID                     = "X-GitHub-Hook-ID"
	HeaderGitHubEvent                      = "X-GitHub-Event"
	HeaderGitHubDelivery                   = "X-GitHub-Delivery"
	HeaderHubSignature                     = "X-Hub-Signature"
	HeaderHubSignature256                  = "X-Hub-Signature-256"
	HeaderGitHubHookInstallationTargetType = "X-GitHub-Hook-Installation-Target-Type"
	HeaderGitHubHookInstallationTargetID   = "X-GitHub-Hook-Installation-Target-ID"
)

type Delivery struct {
	HookId                     string
	Event                      string
	Delivery                   string
	Signature                  string
	Signature256               string
	HookInstallationTargetType string
	HookInstallationTargetId   string
}
