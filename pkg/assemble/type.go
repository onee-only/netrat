package assemble

type AssembleType string

const (
	AssembleTypePlain AssembleType = "plain"
	AssembleTypeHTTP  AssembleType = "http"
)

func (a AssembleType) Valid() bool {
	switch a {
	case AssembleTypePlain, AssembleTypeHTTP:
		return true
	}
	return false
}
