package schego

// Task Interface
type Task interface {
	Exec(id interface{}) error
	Cancel(id interface{}) error
}
