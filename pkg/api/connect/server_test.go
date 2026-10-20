package connectapi_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectinprocess"

	hospitalv1 "ClinicaCoste/gen/go/proto/hospital/v1"
	"ClinicaCoste/gen/go/proto/hospital/v1/hospitalv1connect"
	"ClinicaCoste/internal/hospital"
	"ClinicaCoste/internal/hospital/clock"

	connectapi "ClinicaCoste/pkg/api/connect"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// newTestEnv arma un Hospital con reloj simulado, una habitación (101, cap 2),
// un servidor Connect in-process y devuelve el cliente + el hospital.
//
// IMPORTANTE: no agregamos doctores a propósito, porque Doctor.Attend hace
// time.Sleep(3-6s) y ralentizaría los tests. Sin doctores, dispatch() devuelve
// false y el handler crea el EpisodeRecord directamente.
func newTestEnv(t *testing.T) (*hospital.Hospital, hospitalv1connect.HospitalServiceClient) {
	t.Helper()

	clk := clock.NewSimulatedClock(
		time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
		1.0,
	)
	h, err := hospital.NewHospital("Clínica Coste Test", clk)
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}
	if _, err := h.AddRoom(101, 2); err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	srv := connect.NewServer()
	hospitalv1connect.RegisterHospitalServiceHandler(srv, connectapi.NewHospitalServer(h))

	client := hospitalv1connect.NewHospitalServiceClient(
		connect.NewClient(connectinprocess.New(srv)),
	)
	return h, client
}

// newTestEnvNoRooms arma un Hospital SIN habitaciones, útil para ejercitar
// el camino de ErrNoRoomAvailable.
func newTestEnvNoRooms(t *testing.T) (*hospital.Hospital, hospitalv1connect.HospitalServiceClient) {
	t.Helper()

	clk := clock.NewSimulatedClock(time.Now(), 1.0)
	h, err := hospital.NewHospital("Clínica sin rooms", clk)
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}
	// OJO: no llamamos AddRoom a propósito.

	srv := connect.NewServer()
	hospitalv1connect.RegisterHospitalServiceHandler(srv, connectapi.NewHospitalServer(h))

	client := hospitalv1connect.NewHospitalServiceClient(
		connect.NewClient(connectinprocess.New(srv)),
	)
	return h, client
}

func mustAdmit(t *testing.T, client hospitalv1connect.HospitalServiceClient, name string, age int32, level string) {
	t.Helper()
	_, err := client.AdmitPatient(context.Background(), &hospitalv1.AdmitPatientRequest{
		Name:            name,
		Age:             age,
		NarcolepsyLevel: level,
	})
	if err != nil {
		t.Fatalf("AdmitPatient(%q): %v", name, err)
	}
}

func findPatientID(t *testing.T, h *hospital.Hospital, name string) string {
	t.Helper()
	for _, p := range h.Patients() {
		if p.Name() == name {
			return p.ID()
		}
	}
	t.Fatalf("paciente %q no encontrado", name)
	return ""
}

// -----------------------------------------------------------------------------
// AdmitPatient
// -----------------------------------------------------------------------------

func TestAdmitPatient_Ok(t *testing.T) {
	h, client := newTestEnv(t)

	mustAdmit(t, client, "Ana", 30, "severe")

	patients := h.Patients()
	if len(patients) != 1 {
		t.Fatalf("se esperaba 1 paciente, hay %d", len(patients))
	}
	if patients[0].Name() != "Ana" {
		t.Errorf("name = %q, se esperaba Ana", patients[0].Name())
	}
	if patients[0].Level() != hospital.NarcolepsyLevelSevere {
		t.Errorf("level = %v, se esperaba Severe", patients[0].Level())
	}
}

func TestAdmitPatient_LevelPorDefectoEsMild(t *testing.T) {
	h, client := newTestEnv(t)

	// level desconocido -> el handler cae en Mild
	mustAdmit(t, client, "Beto", 40, "???")

	patients := h.Patients()
	if len(patients) != 1 {
		t.Fatalf("se esperaba 1 paciente, hay %d", len(patients))
	}
	if patients[0].Level() != hospital.NarcolepsyLevelMild {
		t.Errorf("level = %v, se esperaba Mild", patients[0].Level())
	}
}

// -----------------------------------------------------------------------------
// ReportSleepAttack
// -----------------------------------------------------------------------------

func TestReportSleepAttack_AsignaHabitacion(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	resp, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id,
		Location:  "pasillo 2",
	})
	if err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}
	if !resp.GetRoomAssigned() {
		t.Fatal("se esperaba room_assigned=true")
	}
	if resp.GetAssignedRoomNumber() != 101 {
		t.Errorf("assigned_room_number = %d, se esperaba 101", resp.GetAssignedRoomNumber())
	}
}

func TestReportSleepAttack_SinHabitaciones_NoFallaYDevuelveSinRoom(t *testing.T) {
	h, client := newTestEnvNoRooms(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	// El handler ignora ErrNoRoomAvailable y devuelve room_assigned=false
	resp, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id,
		Location:  "pasillo",
	})
	if err != nil {
		t.Fatalf("ReportSleepAttack: %v (se esperaba sin error)", err)
	}
	if resp.GetRoomAssigned() {
		t.Error("se esperaba room_assigned=false")
	}
	if resp.GetAssignedRoomNumber() != 0 {
		t.Errorf("assigned_room_number = %d, se esperaba 0", resp.GetAssignedRoomNumber())
	}
}

func TestReportSleepAttack_PacienteInexistente_NotFound(t *testing.T) {
	_, client := newTestEnv(t)

	_, err := client.ReportSleepAttack(context.Background(), &hospitalv1.ReportSleepAttackRequest{
		PatientId: "no-existe",
		Location:  "pasillo",
	})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Fatalf("code = %v, se esperaba NotFound (err=%v)", got, err)
	}
}

// -----------------------------------------------------------------------------
// AssignRoom
// -----------------------------------------------------------------------------

func TestAssignRoom_PacienteYaEnRoom_FailedPrecondition(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	// ReportSleepAttack ya le asigna la room 101
	if _, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id, Location: "pasillo",
	}); err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}

	// AssignRoom sobre el mismo paciente -> ErrPatientAlreadyInRoom
	_, err := client.AssignRoom(ctx, &hospitalv1.AssignRoomRequest{PatientId: id})
	if err == nil {
		t.Fatal("se esperaba error (ya tiene room)")
	}
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, se esperaba FailedPrecondition (err=%v)", got, err)
	}
}

func TestAssignRoom_PacienteDespierto_FailedPrecondition(t *testing.T) {
	h, client := newTestEnv(t)

	mustAdmit(t, client, "Ana", 30, "mild")
	id := findPatientID(t, h, "Ana")

	_, err := client.AssignRoom(context.Background(), &hospitalv1.AssignRoomRequest{PatientId: id})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, se esperaba FailedPrecondition (err=%v)", got, err)
	}
}

func TestAssignRoom_PacienteInexistente_NotFound(t *testing.T) {
	_, client := newTestEnv(t)

	_, err := client.AssignRoom(context.Background(), &hospitalv1.AssignRoomRequest{PatientId: "nope"})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Fatalf("code = %v, se esperaba NotFound", got)
	}
}

// -----------------------------------------------------------------------------
// AssignRoom -> ErrNoRoomAvailable vía RPC puro
// -----------------------------------------------------------------------------

// TestAssignRoom_SinHabitaciones_ResourceExhausted ejercita el único camino
// RPC que propaga ErrNoRoomAvailable: AssignRoom. ReportSleepAttack lo ignora
// a propósito, así que primero dormimos al paciente (sin room) vía RPC y luego
// pedimos AssignRoom explícitamente.
func TestAssignRoom_SinHabitaciones_ResourceExhausted(t *testing.T) {
	h, client := newTestEnvNoRooms(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	// 1) ReportSleepAttack: el handler ignora ErrNoRoomAvailable, no debe fallar.
	resp, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id,
		Location:  "pasillo",
	})
	if err != nil {
		t.Fatalf("ReportSleepAttack no debía fallar: %v", err)
	}
	if resp.GetRoomAssigned() {
		t.Fatal("room_assigned debía ser false (no hay rooms)")
	}

	// 2) AssignRoom -> ErrNoRoomAvailable -> CodeResourceExhausted
	_, err = client.AssignRoom(ctx, &hospitalv1.AssignRoomRequest{PatientId: id})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if got := connect.CodeOf(err); got != connect.CodeResourceExhausted {
		t.Fatalf("code = %v, se esperaba ResourceExhausted (err=%v)", got, err)
	}
}

// -----------------------------------------------------------------------------
// WakePatient
// -----------------------------------------------------------------------------

func TestWakePatient_Ok(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	if _, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id, Location: "pasillo",
	}); err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}

	if _, err := client.WakePatient(ctx, &hospitalv1.WakePatientRequest{PatientId: id}); err != nil {
		t.Fatalf("WakePatient: %v", err)
	}

	// Tras despertar, no debe estar en el pasillo
	resp, err := client.GetPatientsInHallway(ctx, &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetPatientsInHallway: %v", err)
	}
	if n := len(resp.GetPatients()); n != 0 {
		t.Errorf("se esperaban 0 en pasillo, hay %d", n)
	}
}

func TestWakePatient_YaDespierto_CaeEnInternalPorDefecto(t *testing.T) {
	h, client := newTestEnv(t)

	mustAdmit(t, client, "Ana", 30, "mild")
	id := findPatientID(t, h, "Ana")

	_, err := client.WakePatient(context.Background(), &hospitalv1.WakePatientRequest{PatientId: id})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	// ErrPatientAlreadyAwake no está mapeado en mapDomainError -> CodeInternal.
	// Si en el futuro se agrega a FailedPrecondition, este test fallará y
	// habrá que actualizarlo (es intencional: documenta el comportamiento actual).
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Fatalf("code = %v, se esperaba Internal (err=%v)", got, err)
	}
}

func TestWakePatient_PacienteInexistente_NotFound(t *testing.T) {
	_, client := newTestEnv(t)

	_, err := client.WakePatient(context.Background(), &hospitalv1.WakePatientRequest{PatientId: "nope"})
	if err == nil {
		t.Fatal("se esperaba error")
	}
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Fatalf("code = %v, se esperaba NotFound", got)
	}
}

// -----------------------------------------------------------------------------
// Consultas
// -----------------------------------------------------------------------------

func TestGetPatientsInHallway_Vacio(t *testing.T) {
	_, client := newTestEnv(t)

	resp, err := client.GetPatientsInHallway(context.Background(), &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetPatientsInHallway: %v", err)
	}
	if n := len(resp.GetPatients()); n != 0 {
		t.Errorf("se esperaban 0, hay %d", n)
	}
}

func TestGetPatientsInHallway_ConPaciente(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	// Forzamos el estado dormido sin asignar room: como el hospital tiene
	// habitación 101, ReportSleepAttack le asignaría una. Para tener un paciente
	// dormido en pasillo necesitamos que NO haya rooms disponibles. Usamos el
	// dominio directamente para simular ese estado sin depender del RPC.
	var ana *hospital.Patient
	for _, p := range h.Patients() {
		if p.ID() == id {
			ana = p
			break
		}
	}
	if ana == nil {
		t.Fatal("Ana no encontrada")
	}
	if err := ana.SufferSleepAttack("pasillo"); err != nil {
		t.Fatalf("SufferSleepAttack: %v", err)
	}

	resp, err := client.GetPatientsInHallway(ctx, &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetPatientsInHallway: %v", err)
	}
	if len(resp.GetPatients()) != 1 {
		t.Fatalf("se esperaba 1, hay %d", len(resp.GetPatients()))
	}
	dto := resp.GetPatients()[0]
	if dto.GetId() != id {
		t.Errorf("id = %q, se esperaba %q", dto.GetId(), id)
	}
	if dto.GetName() != "Ana" {
		t.Errorf("name = %q, se esperaba Ana", dto.GetName())
	}
	if !dto.GetIsCrisis() {
		t.Error("is_crisis se esperaba true")
	}
}

func TestGetFloorSnapshot_PisoExistente(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")
	if _, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id, Location: "pasillo",
	}); err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}

	resp, err := client.GetFloorSnapshot(ctx, &hospitalv1.FloorSnapshotRequest{FloorNumber: 1})
	if err != nil {
		t.Fatalf("GetFloorSnapshot: %v", err)
	}
	if resp.GetFloorNumber() != 1 {
		t.Errorf("floor_number = %d, se esperaba 1", resp.GetFloorNumber())
	}
	// Ana está en room 101 de la floor 1 -> debería aparecer
	if len(resp.GetPatients()) != 1 {
		t.Fatalf("se esperaba 1 paciente en floor 1, hay %d", len(resp.GetPatients()))
	}
	if resp.GetPatients()[0].GetAssignedRoom() != 101 {
		t.Errorf("assigned_room = %d, se esperaba 101", resp.GetPatients()[0].GetAssignedRoom())
	}
}

func TestGetFloorSnapshot_PisoInexistente_Vacio(t *testing.T) {
	_, client := newTestEnv(t)

	resp, err := client.GetFloorSnapshot(context.Background(), &hospitalv1.FloorSnapshotRequest{FloorNumber: 99})
	if err != nil {
		t.Fatalf("GetFloorSnapshot: %v", err)
	}
	if n := len(resp.GetPatients()); n != 0 {
		t.Errorf("se esperaban 0, hay %d", n)
	}
}

func TestGetHistory(t *testing.T) {
	h, client := newTestEnv(t)
	ctx := context.Background()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")
	if _, err := client.ReportSleepAttack(ctx, &hospitalv1.ReportSleepAttackRequest{
		PatientId: id, Location: "pasillo 3",
	}); err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}

	resp, err := client.GetHistory(ctx, &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(resp.GetEpisodes()) != 1 {
		t.Fatalf("se esperaba 1 episodio, hay %d", len(resp.GetEpisodes()))
	}
	ep := resp.GetEpisodes()[0]
	if ep.GetPatientId() != id {
		t.Errorf("patient_id = %q, se esperaba %q", ep.GetPatientId(), id)
	}
	if ep.GetPatientName() != "Ana" {
		t.Errorf("patient_name = %q, se esperaba Ana", ep.GetPatientName())
	}
	if ep.GetLocation() != "pasillo 3" {
		t.Errorf("location = %q, se esperaba 'pasillo 3'", ep.GetLocation())
	}
	if ep.GetAttendingDoctor() != "unassigned" {
		t.Errorf("attending_doctor = %q, se esperaba 'unassigned'", ep.GetAttendingDoctor())
	}
	if ep.GetAssignedRoom() != "room 101" {
		t.Errorf("assigned_room = %q, se esperaba 'room 101'", ep.GetAssignedRoom())
	}
}

func TestGetStaff_Vacio(t *testing.T) {
	_, client := newTestEnv(t)

	resp, err := client.GetStaff(context.Background(), &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetStaff: %v", err)
	}
	if n := len(resp.GetStaff()); n != 0 {
		t.Errorf("se esperaban 0, hay %d", n)
	}
}

func TestGetStaff_ConDoctor(t *testing.T) {
	h, client := newTestEnv(t)

	doc := hospital.NewDoctor("Dr. House", 50, "Diagnóstico")
	if err := h.HireDoctor(doc); err != nil {
		t.Fatalf("HireDoctor: %v", err)
	}

	resp, err := client.GetStaff(context.Background(), &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("GetStaff: %v", err)
	}
	if len(resp.GetStaff()) != 1 {
		t.Fatalf("se esperaba 1 staff, hay %d", len(resp.GetStaff()))
	}
	if resp.GetStaff()[0].GetRole() != "Doctor" {
		t.Errorf("role = %q, se esperaba Doctor", resp.GetStaff()[0].GetRole())
	}
	if resp.GetStaff()[0].GetName() != "Dr. House" {
		t.Errorf("name = %q, se esperaba 'Dr. House'", resp.GetStaff()[0].GetName())
	}
}

// -----------------------------------------------------------------------------
// Streaming: recepción de eventos
// -----------------------------------------------------------------------------

func TestStreamEpisodes_RecibeEvento(t *testing.T) {
	h, client := newTestEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mustAdmit(t, client, "Ana", 30, "severe")
	id := findPatientID(t, h, "Ana")

	stream, err := client.StreamEpisodes(ctx, &hospitalv1.Empty{})
	if err != nil {
		t.Fatalf("StreamEpisodes: %v", err)
	}
	defer stream.Close()

	// Damos un pequeño margen para que el server se suscriba antes de emitir.
	time.Sleep(50 * time.Millisecond)

	if _, err := client.ReportSleepAttack(context.Background(), &hospitalv1.ReportSleepAttackRequest{
		PatientId: id, Location: "pasillo",
	}); err != nil {
		t.Fatalf("ReportSleepAttack: %v", err)
	}

	evt, err := stream.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if evt.GetPatientId() != id {
		t.Errorf("patient_id = %q, se esperaba %q", evt.GetPatientId(), id)
	}
	if evt.GetLocation() != "pasillo" {
		t.Errorf("location = %q, se esperaba pasillo", evt.GetLocation())
	}
	if evt.GetAssignedRoom() != "room 101" {
		t.Errorf("assigned_room = %q, se esperaba 'room 101'", evt.GetAssignedRoom())
	}
}

// -----------------------------------------------------------------------------
// Streaming: cancelación del cliente
// -----------------------------------------------------------------------------

// TestStreamEpisodes_CancelacionCliente verifica que al cancelar el contexto
// del cliente:
//   - stream.Receive() devuelve un error de cancelación (CodeCanceled o
//     context.Canceled), no un mensaje espurio.
//   - El server no queda colgado: el handler hace defer cancel(), que cierra
//     el canal de suscripción y lo elimina de h.subs.
//
// Como h.subs es privado, la verificación de limpieza se hace en un test del
// paquete hospital (ver nota al final del archivo).
func TestStreamEpisodes_CancelacionCliente(t *testing.T) {
	h, client := newTestEnv(t)

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := client.StreamEpisodes(ctx, &hospitalv1.Empty{})
	if err != nil {
		cancel()
		t.Fatalf("StreamEpisodes: %v", err)
	}

	// Damos margen para que el server registre la suscripción.
	time.Sleep(50 * time.Millisecond)

	// Cancelamos el contexto del cliente.
	cancel()

	// Receive debe fallar con cancelación. Puede devolver:
	//   - connect.Error con CodeCanceled, o
	//   - context.Canceled directamente.
	// Aceptamos ambos, porque depende de cómo Connect v2 envuelva el error.
	evt, err := stream.Receive()
	if err == nil {
		t.Fatalf("Receive debía fallar tras cancelar, recibió: %+v", evt)
	}

	canceled := errors.Is(err, context.Canceled) ||
		connect.CodeOf(err) == connect.CodeCanceled
	if !canceled {
		t.Fatalf("se esperaba cancelación (context.Canceled o CodeCanceled), "+
			"se obtuvo code=%v err=%v", connect.CodeOf(err), err)
	}

	// Cerrar el stream es idempotente y no debe paniquear.
	_ = stream.Close()

	// Verificación indirecta de que el server no quedó colgado:
	// publicamos un evento y el hospital debe poder seguir operando.
	// Si el handler hubiera dejado el canal abierto sin drenar, publish no
	// bloquearía igual (select con default), pero al menos confirmamos que
	// el flujo sigue vivo.
	mustAdmit(t, client, "Beto", 40, "severe")
	idBeto := findPatientID(t, h, "Beto")
	if _, err := client.ReportSleepAttack(context.Background(), &hospitalv1.ReportSleepAttackRequest{
		PatientId: idBeto, Location: "pasillo",
	}); err != nil {
		t.Fatalf("ReportSleepAttack tras cancelar stream: %v", err)
	}
}

// TestStreamEpisodes_CancelacionNoBloqueaPublish verifica que, tras cancelar
// varios streams, el hospital sigue publicando eventos sin bloquearse ni
// paniquear por enviar a canales cerrados (publish usa select con default y
// cancel cierra el canal bajo subMu, por lo que publish no ve canales cerrados).
func TestStreamEpisodes_CancelacionNoBloqueaPublish(t *testing.T) {
	h, client := newTestEnv(t)

	// Abrimos y cancelamos varios streams para estresar suscripción/cancelación.
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := client.StreamEpisodes(ctx, &hospitalv1.Empty{})
		if err != nil {
			cancel()
			t.Fatalf("StreamEpisodes[%d]: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
		cancel()
		_ = stream.Close()
	}

	// Publicar un evento no debe bloquear ni paniquear.
	done := make(chan struct{})
	go func() {
		defer close(done)
		mustAdmit(t, client, "Ana", 30, "severe")
		id := findPatientID(t, h, "Ana")
		_, _ = client.ReportSleepAttack(context.Background(), &hospitalv1.ReportSleepAttackRequest{
			PatientId: id, Location: "pasillo",
		})
	}()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("ReportSleepAttack se bloqueó tras cancelar streams")
	}
}
