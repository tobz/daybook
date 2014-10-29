package daybook

type Archive interface {
	Extract(rootDir string) error
}
