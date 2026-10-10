package hospital

import (
	"errors"
	"testing"
)

// Comprobación en tiempo de compilación de que los tres tipos satisfacen
// Attender — exactamente lo que pide la sección 4.2 del enunciado.
var (
	_ Attender = (*Doctor)(nil)
	_ Attender = (*Orderly)(nil)
	_ Attender = (*Nurse)(nil)
)

// TestAttender_Polymorphism mete Doctor, Orderly y Nurse en el mismo
// []Attender y los despacha sin switch de tipos concretos: es el punto que
// el enunciado marca como "score cero si no se usa así".
func TestAttender_Polymorphism(t *testing.T) {
	staff := []Attender{
		NewDoctor("Karen Ospina", 38, "Neurology"),
		NewOrderly("Camilo Pérez", 25),
		NewNurse("Andrea Gómez", 30),
	}

	for _, a := range staff {
		p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)
		_ = p.SufferSleepAttack("cafeteria")

		e, err := a.Attend(p, "cafeteria")
		if err != nil {
			t.Fatalf("%s.Attend: %v", a.Name(), err)
		}
		if e.AttendingDoctor().ID() != a.ID() {
			t.Errorf("%s: episode attended by wrong ID", a.Name())
		}
		if p.AssignedDoctor() == nil || p.AssignedDoctor().ID() != a.ID() {
			t.Errorf("%s: patient not linked back to the attender", a.Name())
		}
	}
}

func TestOrderlyAttend(t *testing.T) {
	o := NewOrderly("Camilo", 25)
	p := NewPatient("Wilfrido", 40, NarcolepsyLevelSevere)

	if _, err := o.Attend(p, "x"); !errors.Is(err, ErrPatientNotAsleep) {
		t.Fatalf("awake patient: got %v", err)
	}
	_ = p.SufferSleepAttack("hallway 2")

	e, err := o.Attend(p, "hallway 2")
	if err != nil {
		t.Fatalf("Attend: %v", err)
	}
	if o.CanDiagnose() {
		t.Error("an orderly can never diagnose")
	}
	if got := o.PatientsMoved(); len(got) != 1 || got[0].ID() != p.ID() {
		t.Errorf("orderly should track who it moved: %v", got)
	}
	if e.Location() != "hallway 2" {
		t.Errorf("unexpected episode: %v", e)
	}
}

func TestNurseAttend(t *testing.T) {
	n := NewNurse("Andrea", 30)
	p := NewPatient("Yeimy", 29, NarcolepsyLevelModerate)

	if _, err := n.Attend(nil, "x"); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil patient: got %v", err)
	}
	_ = p.SufferSleepAttack("radiology queue")

	if _, err := n.Attend(p, "radiology queue"); err != nil {
		t.Fatalf("Attend: %v", err)
	}
	if n.CanDiagnose() {
		t.Error("a nurse can never diagnose")
	}
	if got := n.Stabilized(); len(got) != 1 || got[0].ID() != p.ID() {
		t.Errorf("nurse should track who it stabilized: %v", got)
	}
}
