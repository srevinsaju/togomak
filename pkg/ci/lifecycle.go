package ci

type LifecycleType int64

// Lifecycle is inspired from the Maven build lifecycles"
//   validate - validate the project is correct and all necessary information is available
//   compile - compile the source code of the project
//   test - test the compiled source code using a suitable unit testing framework. These tests should not require the code be packaged or deployed
//   package - take the compiled code and package it in its distributable format, such as a JAR.
//   verify - run any checks on results of integration tests to ensure quality criteria are met
//   install - install the package into the local repository, for use as a dependency in other projects locally
//   deploy - done in the build environment, copies the final package to the remote repository for sharing with other developers and projects.

const (
	LifecycleDefault LifecycleType = iota

	LifecycleValidate
	LifecycleCompile
	LifecycleTest
	LifecyclePackage
	LifecycleVerify
	LifecycleInstall
	LifecycleDeploy

	LifecycleInvalid = -1
)

func (ty LifecycleType) String() string {
	switch ty {
	case LifecycleDefault:
		return "default"
	case LifecycleValidate:
		return "validate"
	case LifecycleCompile:
		return "compile"
	case LifecycleTest:
		return "test"
	case LifecyclePackage:
		return "package"
	case LifecycleVerify:
		return "verify"
	case LifecycleInstall:
		return "install"
	case LifecycleDeploy:
		return "deploy"
	}
	panic("invalid lifecycle type")
}

func LifecycleUnmarshall(v string) (LifecycleType, bool) {
	lifecycle := []LifecycleType{
		LifecycleDefault,
		LifecycleValidate,
		LifecycleCompile,
		LifecycleTest,
		LifecyclePackage,
		LifecycleVerify,
		LifecycleDeploy,
	}
	for _, ly := range lifecycle {
		if ly.String() == v {
			return ly, true
		}
	}
	return LifecycleInvalid, false

}
