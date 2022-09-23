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

type Name int

const (
	John Name = iota
	Jenny
	Bob
)

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
