package fireblazer

type ServiceMeta struct {
	Title   string
	Summary string
}

// TODO: Figure out what to do with these stubs. Go all in with separate initialization and one master init in embed.go? Or all in on embed.go?

var ApiMetadata map[string]ServiceMeta // a shell of what was :3 (check embed.go to understand where it comes from, it's all initialized there)
