package testdata

// Animal is a struct with a name
type Animal struct {
	Name string
}

// Name of the Animal
func (a *Animal) Name() string {
	return a.Name
}

// Person is a struct with a name
type Person struct {
	Name string
}

// Name of the Person
func (p *Person) Name() string {
	return a.Name
}

// GetName of a namer
func GetName(namer interface{ Name() string }) string {
	return namer.Name()
}
