package testdata

import "fmt"

type (
	// Age of a Person
	Age int
	// Weight of a Person
	Weight int
)

// Animal is a struct with a name
type Animal struct {
	name string
}

// Name of the Animal
func (a *Animal) Name() string {
	return a.name
}

// Name of a Person
type Name int

// IsValid Name
func (n Name) IsValid() bool {
	switch n {
	case John,
		Jenny,
		Bob:
		return true
	default:
		return false
	}
}

// String of a Name
func (n Name) String() string {
	switch n {
	case John:
		return "John"
	case Jenny:
		return "Jenny"
	case Bob:
		return "Bob"
	}
}

// String version of Car name in a different file
func (c Car) String() string {
	return c.Name()
}

// Namer is something implementing a Name() string method
type Namer interface {
	Name() string
}

// Person is a struct with a name
type Person struct {
	name string
}

// Name of the Person
func (p *Person) Name() string {
	return p.name
}

// GetName of a namer
func GetName(n Namer) string {
	return n.Name()
}

// PrintName of a namer
func PrintName(n Namer) {
	fmt.Println(n)
}

const (
	// John is someone named "John"
	John Name = iota + 1
	// Jenny is someone named "Jenny"
	Jenny
	// Bob is someone named "Bob"
	Bob
)
