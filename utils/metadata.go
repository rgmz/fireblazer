package fireblazer

type ServiceMeta struct {
	Title   string
	Summary string
}

var ApiMetadata map[string]ServiceMeta // a shell of what was :3 (check embed.go to understand where it comes from, it's all initialized there)
