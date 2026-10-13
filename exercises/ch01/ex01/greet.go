package greet

// Greet returns a greeting for the given name.
func Greet(name string) string {
	if len(name) > 0 {
		return "Hello, " + name + "!"
	}
	return "Hello, World!"
}
