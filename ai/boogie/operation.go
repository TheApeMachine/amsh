package boogie

type Operation struct {
	Name     string
	Outcomes []string  // e.g., ["next", "back", "cancel"]
	Label    string    // Optional label for referencing
	Children []Program // For nested operations
}