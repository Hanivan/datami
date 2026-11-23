package nautiljon

const (
	OpGet        = "get"
	OpSearch     = "search"
	OpGetByMalID = "getbymalid"
)

const (
	defaultBaseURL = "https://www.nautiljon.com"
	defaultRate    = 1
)

var nautiljonStatusMap = map[string]string{
	"en cours": "ongoing",
	"terminé":  "finished",
}
