package connectapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect/v2"

	hospitalv1 "ClinicaCoste/gen/go/proto/hospital/v1"
	"ClinicaCoste/gen/go/proto/hospital/v1/hospitalv1connect"
	hospital "ClinicaCoste/internal/hospital"
)

// HospitalServer implementa el servicio gRPC/Connect definido en el .proto
type HospitalServer struct {
	hospitalv1connect.UnimplementedHospitalServiceHandler
	h *hospital.Hospital
}

func NewHospitalServer(h *hospital.Hospital) *HospitalServer {
	return &HospitalServer{h: h}
}

// ---------------------------------------------------------------------------
// 1. GESTIÓN Y SIMULACIÓN
// ---------------------------------------------------------------------------

func (s *HospitalServer) AdmitPatient(
	ctx context.Context,
	req *hospitalv1.AdmitPatientRequest,
) (*hospitalv1.AdmitPatientResponse, error) {
	var level hospital.NarcolepsyLevel
	switch req.NarcolepsyLevel {
	case "mild", "Mild", "1":
		level = hospital.NarcolepsyLevelMild
	case "moderate", "Moderate", "2":
		level = hospital.NarcolepsyLevelModerate
	case "severe", "Severe", "3":
		level = hospital.NarcolepsyLevelSevere
	default:
		level = hospital.NarcolepsyLevelMild
	}

	p := hospital.NewPatient(req.Name, int(req.Age), level)

	err := s.h.AdmitPatient(p)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &hospitalv1.AdmitPatientResponse{}, nil
}

func (s *HospitalServer) ReportSleepAttack(
	ctx context.Context,
	req *hospitalv1.ReportSleepAttackRequest,
) (*hospitalv1.ReportSleepAttackResponse, error) {
	patients := s.h.Patients()
	var targetPatient *hospital.Patient
	for _, p := range patients {
		if p.ID() == req.PatientId {
			targetPatient = p
			break
		}
	}
	if targetPatient == nil {
		return nil, mapDomainError(hospital.ErrPatientNotFound)
	}

	room, err := s.h.ReportSleepAttack(targetPatient, req.Location)
	if err != nil && !errors.Is(err, hospital.ErrNoRoomAvailable) {
		return nil, mapDomainError(err)
	}

	roomNumber := int32(0)
	roomAssigned := false
	if room != nil {
		roomNumber = int32(room.Number())
		roomAssigned = true
	}

	return &hospitalv1.ReportSleepAttackResponse{
		AssignedRoomNumber: roomNumber,
		RoomAssigned:       roomAssigned,
	}, nil
}

func (s *HospitalServer) AssignRoom(
	ctx context.Context,
	req *hospitalv1.AssignRoomRequest,
) (*hospitalv1.AssignRoomResponse, error) {
	patients := s.h.Patients()
	var targetPatient *hospital.Patient
	for _, p := range patients {
		if p.ID() == req.PatientId {
			targetPatient = p
			break
		}
	}
	if targetPatient == nil {
		return nil, mapDomainError(hospital.ErrPatientNotFound)
	}

	room, err := s.h.AssignRoom(targetPatient)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &hospitalv1.AssignRoomResponse{
		RoomNumber: int32(room.Number()),
	}, nil
}

func (s *HospitalServer) WakePatient(
	ctx context.Context,
	req *hospitalv1.WakePatientRequest,
) (*hospitalv1.WakePatientResponse, error) {
	patients := s.h.Patients()
	var targetPatient *hospital.Patient
	for _, p := range patients {
		if p.ID() == req.PatientId {
			targetPatient = p
			break
		}
	}
	if targetPatient == nil {
		return nil, mapDomainError(hospital.ErrPatientNotFound)
	}

	if err := s.h.WakePatient(targetPatient); err != nil {
		return nil, mapDomainError(err)
	}

	return &hospitalv1.WakePatientResponse{}, nil
}

// ---------------------------------------------------------------------------
// 2. CONSULTAS DE ESTADO Y VISTAS
// ---------------------------------------------------------------------------

func (s *HospitalServer) GetPatientsInHallway(
	ctx context.Context,
	req *hospitalv1.Empty,
) (*hospitalv1.PatientsInHallwayResponse, error) {
	hallwayPatients := s.h.PatientsInHallway()
	dtos := make([]*hospitalv1.PatientDto, len(hallwayPatients))

	for i, p := range hallwayPatients {
		roomNum := int32(0)
		if p.AssignedRoom() != nil {
			roomNum = int32(p.AssignedRoom().Number())
		}
		dtos[i] = &hospitalv1.PatientDto{
			Id:           p.ID(),
			Name:         p.Name(),
			State:        p.State().String(),
			Location:     p.Location(),
			AssignedRoom: roomNum,
			IsCrisis:     p.IsAsleepInHallway(),
		}
	}

	return &hospitalv1.PatientsInHallwayResponse{Patients: dtos}, nil
}

func (s *HospitalServer) GetFloorSnapshot(
	ctx context.Context,
	req *hospitalv1.FloorSnapshotRequest,
) (*hospitalv1.FloorSnapshotResponse, error) {
	views := s.h.FloorSnapshot(int(req.FloorNumber))
	dtos := make([]*hospitalv1.PatientDto, len(views))

	for i, v := range views {
		dtos[i] = &hospitalv1.PatientDto{
			Id:           v.ID,
			Name:         v.Name,
			State:        v.State,
			Location:     v.Location,
			AssignedRoom: int32(v.AssignedRoom),
			IsCrisis:     v.IsCrisis,
		}
	}

	return &hospitalv1.FloorSnapshotResponse{
		FloorNumber: req.FloorNumber,
		Patients:    dtos,
	}, nil
}

func (s *HospitalServer) GetHistory(
	ctx context.Context,
	req *hospitalv1.Empty,
) (*hospitalv1.HistoryResponse, error) {
	history := s.h.History()
	dtos := make([]*hospitalv1.EpisodeEventDto, len(history))

	for i, e := range history {
		patientID := ""
		patientName := ""
		if e.Patient() != nil {
			patientID = e.Patient().ID()
			patientName = e.Patient().Name()
		}
		docName := "unassigned"
		if e.AttendingDoctor() != nil {
			docName = e.AttendingDoctor().Name()
		}
		roomStr := "hallway"
		if e.AssignedRoom() != nil {
			roomStr = fmt.Sprintf("room %d", e.AssignedRoom().Number())
		}

		dtos[i] = &hospitalv1.EpisodeEventDto{
			Id:              e.ID(),
			Timestamp:       e.Timestamp().Format(time.RFC3339),
			PatientId:       patientID,
			PatientName:     patientName,
			AttendingDoctor: docName,
			Location:        e.Location(),
			AssignedRoom:    roomStr,
		}
	}

	return &hospitalv1.HistoryResponse{Episodes: dtos}, nil
}

func (s *HospitalServer) GetStaff(
	ctx context.Context,
	req *hospitalv1.Empty,
) (*hospitalv1.StaffResponse, error) {
	staffList := s.h.Staff()
	dtos := make([]*hospitalv1.StaffMemberDto, len(staffList))

	for i, st := range staffList {
		role := "Staff"
		if _, ok := st.(*hospital.Doctor); ok {
			role = "Doctor"
		}
		dtos[i] = &hospitalv1.StaffMemberDto{
			Id:   st.ID(),
			Name: st.Name(),
			Role: role,
		}
	}

	return &hospitalv1.StaffResponse{Staff: dtos}, nil
}

// ---------------------------------------------------------------------------
// 3. SERVER-STREAMING RPC (TABLERO EN VIVO)
// ---------------------------------------------------------------------------

func (s *HospitalServer) StreamEpisodes(
	ctx context.Context,
	req *hospitalv1.Empty,
	stream hospitalv1connect.HospitalServiceStreamEpisodesServerStream,
) error {
	ch, cancel := s.h.Subscribe(50)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case episode, ok := <-ch:
			if !ok {
				return nil
			}

			patientID, patientName := "", ""
			if episode.Patient() != nil {
				patientID = episode.Patient().ID()
				patientName = episode.Patient().Name()
			}
			docName := "unassigned"
			if episode.AttendingDoctor() != nil {
				docName = episode.AttendingDoctor().Name()
			}
			roomStr := "hallway"
			if episode.AssignedRoom() != nil {
				roomStr = fmt.Sprintf("room %d", episode.AssignedRoom().Number())
			}

			dto := &hospitalv1.EpisodeEventDto{
				Id:              episode.ID(),
				Timestamp:       episode.Timestamp().Format(time.RFC3339),
				PatientId:       patientID,
				PatientName:     patientName,
				AttendingDoctor: docName,
				Location:        episode.Location(),
				AssignedRoom:    roomStr,
			}

			if err := stream.Send(dto); err != nil {
				return err
			}
		}
	}
}

func (s *HospitalServer) GetAllPatients(
	ctx context.Context,
	req *hospitalv1.GetAllPatientsRequest,
) (*hospitalv1.GetAllPatientsResponse, error) {
	patients := s.h.Patients()
	dtos := make([]*hospitalv1.PatientDto, len(patients))

	for i, p := range patients {
		roomNum := int32(0)
		if p.AssignedRoom() != nil {
			roomNum = int32(p.AssignedRoom().Number())
		}
		dtos[i] = &hospitalv1.PatientDto{
			Id:              p.ID(),
			Name:            p.Name(),
			State:           p.State().String(),
			Location:        p.Location(),
			AssignedRoom:    roomNum,
			IsCrisis:        p.IsAsleepInHallway(),
			NarcolepsyLevel: p.Level().String(),
		}
	}

	return &hospitalv1.GetAllPatientsResponse{Patients: dtos}, nil
}

func (s *HospitalServer) GetRooms(
	ctx context.Context,
	req *hospitalv1.Empty,
) (*hospitalv1.GetRoomsResponse, error) {
	rooms := s.h.Rooms()
	dtos := make([]*hospitalv1.RoomDto, len(rooms))
	for i, r := range rooms {
		dtos[i] = &hospitalv1.RoomDto{
			Number:   int32(r.Number()),
			Capacity: int32(r.Capacity()),
			State:    r.State().String(),
		}
	}
	return &hospitalv1.GetRoomsResponse{Rooms: dtos}, nil
}

func (s *HospitalServer) GetSeverityReport(ctx context.Context, req *hospitalv1.Empty) (*hospitalv1.SeverityReportResponse, error) {
	report := s.h.SevereReportOrdered()
	entries := make([]*hospitalv1.SeverityReportEntry, len(report))
	for i, r := range report {
		entries[i] = &hospitalv1.SeverityReportEntry{
			PatientId:    r.Patient.ID(),
			PatientName:  r.Patient.Name(),
			EpisodeCount: int32(r.Episodes),
		}
	}
	return &hospitalv1.SeverityReportResponse{Entries: entries}, nil
}

func (s *HospitalServer) GetDoctorStats(ctx context.Context, req *hospitalv1.Empty) (*hospitalv1.DoctorStatsResponse, error) {
	staffList := s.h.Staff()
	entries := make([]*hospitalv1.DoctorStatEntry, 0)

	for _, st := range staffList {
		role := "Staff"
		episodesCount := 0

		if doc, ok := st.(*hospital.Doctor); ok {
			role = "Doctor"
			episodesCount = len(doc.MyEpisodes()) // <-- Usamos la función exacta del PDF
		}

		entries = append(entries, &hospitalv1.DoctorStatEntry{
			DoctorId:     st.ID(),
			DoctorName:   st.Name(),
			Role:         role,
			EpisodeCount: int32(episodesCount),
		})
	}
	return &hospitalv1.DoctorStatsResponse{Entries: entries}, nil
}

// ---------------------------------------------------------------------------
// MAPEO CENTRALIZADO DE ERRORES DEL DOMINIO A CONNECT
// ---------------------------------------------------------------------------

func mapDomainError(err error) *connect.Error {
	switch {
	case errors.Is(err, hospital.ErrPatientNotFound), errors.Is(err, hospital.ErrDoctorNotFound):
		return connect.NewError(connect.CodeNotFound, err.Error())
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		return connect.NewError(connect.CodeResourceExhausted, err.Error())
	case errors.Is(err, hospital.ErrPatientAlreadyAdmitted), errors.Is(err, hospital.ErrDoctorAlreadyHired), errors.Is(err, hospital.ErrRoomAlreadyRegistered):
		return connect.NewError(connect.CodeAlreadyExists, err.Error())
	case errors.Is(err, hospital.ErrNilPatient), errors.Is(err, hospital.ErrNilDoctor), errors.Is(err, hospital.ErrNilAttender):
		return connect.NewError(connect.CodeInvalidArgument, err.Error())
	case errors.Is(err, hospital.ErrPatientNotAsleep), errors.Is(err, hospital.ErrPatientAlreadyInRoom), errors.Is(err, hospital.ErrNoPatientWaiting):
		return connect.NewError(connect.CodeFailedPrecondition, err.Error())
	default:
		return connect.NewError(connect.CodeInternal, err.Error())
	}
}
