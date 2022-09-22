package testdata

import "fmt"

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

type Namer interface {
	Name() string
}

// GetName of a namer
func GetName(n Namer) string {
	return n.Name()
}

// PrintName of a namer
func PrintName(n Namer) {
	fmt.Println(n)
}
