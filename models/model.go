package models

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Main            string            `json:"main"`
	License         string            `json:"license"`
	Scripts         map[string]string `json:"scripts"`
	Author          string            `json:"author"`
	Repository      *Repository       `json:"repository,omitempty"`
	Description     string            `json:"description"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

type PackageData struct {
	Name     string                  `json:"name"`
	Versions map[string]*VersionData `json:"versions"`
}

type VersionData struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	Dist         *Dist             `json:"dist,omitempty"`
}

type Dist struct {
	Tarball   string `json:"tarball"`
	Integrity string `json:"integrity"`
}
