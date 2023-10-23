package ci

const BuilderBlock = "togomak"

// Builder block, also known as the togomak block as defined as the BuilderBlock constant
// determines if a *.hcl file and the directory in which it is placed is a togomak pipeline.
// If this file was not present, the directory is not a togomak pipeline.
//
// Builder accepts a single argument, which is the version of the pipeline.
// The version is used to determine the behavior of the pipeline configuration, i.e. the *.hcl file.
// Supported values of Version are 1 and 2. The Conductor will fail abruptly if the version is not
// supported.
type Builder struct {
	Version int `hcl:"version" json:"version"`
}
