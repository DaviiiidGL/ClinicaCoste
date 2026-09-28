package hospital

import "testing"

func TestNewPerson(t *testing.T) {
	a := NewPerson("Karen Ospina", 38)
	b := NewPerson("Karen Ospina", 38)

	if a.ID() == "" || a.Name() != "Karen Ospina" || a.Age() != 38 {
		t.Errorf("unexpected person: %+v", a)
	}
	if a.ID() == b.ID() {
		t.Error("two people must never share an ID")
	}
}

func TestPersonIsPromotedThroughEmbedding(t *testing.T) {
	p := NewPatient("Yeimy", 29, NarcolepsyLevelSevere)
	d := NewDoctor("Karen", 38, "Neurology")

	if p.Name() != "Yeimy" || p.Age() != 29 || p.ID() == "" {
		t.Errorf("Patient did not inherit Person: %v", p)
	}
	if d.Name() != "Karen" || d.Age() != 38 || d.ID() == "" {
		t.Errorf("Doctor did not inherit Person: %+v", d)
	}
}

// Comprobación en tiempo de compilación: por los getters con receptor por
// valor, tanto Doctor como *Doctor tienen ID() y Name().
func TestDoctorValueAndPointerHaveIdentityMethods(t *testing.T) {
	type identity interface {
		ID() string
		Name() string
	}
	var _ identity = &Doctor{}
	var _ identity = &Doctor{}
}

func TestShortID(t *testing.T) {
	if got := shortID("123456789abc"); got != "12345678" {
		t.Errorf("got %q", got)
	}
	if got := shortID("abc"); got != "abc" {
		t.Errorf("short IDs must be untouched, got %q", got)
	}
}
