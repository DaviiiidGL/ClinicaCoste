package hospital

import "uuid"

type Person struct {
	id   string
	name string
	age  int
}

func NewPerson(name string, age int) *Person {
	return &Person{
		id:   uuid.New().String(),
		name: name,
		age:  age,
	}
}

func (p *Person) ID() string {
	return p.id
}

func (p *Person) Name() string {
	return p.name
}

func (p *Person) Age() int {
	return p.age
}

func (p *Person) SetAge(age int) {
	if age > p.age {
		p.age = age
	}
}
