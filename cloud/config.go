package cloud

type ConfigOption interface {
	Set(*settings)
}

type settings struct {
	gcp gcpClientConfig
}

type project string

func SetProject(p string) ConfigOption {
	return project(p)
}

func (p project) Set(s *settings) {
	s.gcp.Project = string(p)
}
