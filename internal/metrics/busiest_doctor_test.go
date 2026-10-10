package metrics_test

import (
	"testing"

	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/metrics"
)

func TestBusiestDoctor(t *testing.T) {
	h := newHospital(t, 1, 1, 1)
	karen := hospital.NewDoctor("Karen Ospina", 38, "Neurology")
	luis := hospital.NewDoctor("Luis Mendoza", 45, "Sleep medicine")
	_ = h.HireDoctor(karen)
	_ = h.HireDoctor(luis)

	// El despacho es round-robin: Karen, Luis, Karen.
	for _, name := range []string{"A", "B", "C"} {
		attack(t, h, admit(t, h, name), "cafeteria")
	}

	doc, n, ok := metrics.BusiestDoctor(h)

	if !ok || doc != karen || n != 2 {
		t.Errorf("want Karen with 2, got (%v, %d, %v)", doc, n, ok)
	}
}

func TestBusiestDoctor_TieBreaksByName(t *testing.T) {
	h := newHospital(t, 1, 1)
	zoe := hospital.NewDoctor("Zoe", 40, "Neurology")
	ana := hospital.NewDoctor("Ana", 40, "Neurology")
	_ = h.HireDoctor(zoe) // contratada primero: atiende primero
	_ = h.HireDoctor(ana)
	attack(t, h, admit(t, h, "A"), "x")
	attack(t, h, admit(t, h, "B"), "y")

	doc, _, ok := metrics.BusiestDoctor(h)

	if !ok || doc != ana {
		t.Errorf("tie must resolve alphabetically (Ana), got %v", doc)
	}
}

func TestBusiestDoctor_NoneAttendedYet(t *testing.T) {
	h := newHospital(t, 1)
	if _, _, ok := metrics.BusiestDoctor(h); ok {
		t.Error("no doctors: ok must be false")
	}

	_ = h.HireDoctor(hospital.NewDoctor("Karen", 38, "Neurology"))
	if _, _, ok := metrics.BusiestDoctor(h); ok {
		t.Error("hired but nobody attended yet: ok must be false")
	}
}
