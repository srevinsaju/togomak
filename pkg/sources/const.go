package sources

const (
	// TypeGit specifies the git source type
	TypeGit = "git"

	// TypeFile specifies the file source type
	TypeFile = "file"
	// TODO: implement http, https, etc.
)

const (
	// TypeGitConfigCloneDepth allows users to configure the clone depth
	TypeGitConfigCloneDepth = "clone_depth"

	// TypeGitConfigSingleBranch allows users to configure if only a single branch is cloned
	TypeGitConfigSingleBranch = "single_branch"

	// TypeGitConfigInsecureSkipTLS allows users to configure if TLS is skipped
	TypeGitConfigInsecureSkipTLS = "insecure_skip_tls"

	// TypeGitConfigDefaultsCloneDepth gives the default clone depth
	TypeGitConfigDefaultsCloneDepth = "1"

	TypeGitConfigReferenceName = "reference_name"
)
