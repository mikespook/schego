package schego

// Task Interface
type Task interface {
	Exec() error
	Cancel() error
}
