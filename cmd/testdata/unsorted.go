package testdata

import "fmt"

// Person is a struct with a name
type Person struct {
	name string
}

// Name of the Person
func (p *Person) Name() string {
	return p.name
}

// Animal is a struct with a name
type Animal struct {
	name string
}

// Name of the Animal
func (a *Animal) Name() string {
	return a.name
}

// Namer is something implementing a Name() string method
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

type (
	// Name of a Person
	Name int
	// Weight of a Person
	Weight int
	// Age of a Person
	Age int
)

const (
	// John is someone named "John"
	John Name = iota + 1
	// Jenny is someone named "Jenny"
	Jenny
	// Bob is someone named "Bob"
	Bob
)

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
