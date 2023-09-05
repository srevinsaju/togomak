package ci

type Import struct {
	Source string `hcl:"source" json:"source"`
}

type Imports []Import
