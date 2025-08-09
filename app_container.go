package container

type AppContainer interface {
	Container

	// SetModuleGlobalConfig adds static configurations of modules in to the container.
	SetModuleGlobalConfig(configs ...ModuleConfig) error

	// Start starts modules iteratively in the order they are provided.
	//
	// This is done by invoking the Run() method of each module.
	// Before Run() is called readiness of each module is verified using Ready().
	Start(modules ...string)

	// Shutdown gracefully shuts down modules in the order they are provided.
	Shutdown(modules ...string)
}

// Runnable interface is used for modules that needs a runnable process.
type Runnable interface {
	Initable

	// Run starts the module.
	Run() error
}

// Stoppable interface is used to gracefully stop running modules.
type Stoppable interface {
	Stop() error
}

type Validator interface {
	Validator() error
}
