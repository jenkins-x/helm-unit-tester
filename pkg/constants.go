package pkg

const (
	// resourcesSeparator the separator between k8s resources in the output of helm template
	resourcesSeparator = "---"

	// DefaultDirWritePermissions for directories
	DefaultDirWritePermissions = 0760

	// DefaultFileWritePermissions default permissions when creating a file
	DefaultFileWritePermissions = 0644
)
