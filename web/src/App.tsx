// web/src/App.tsx
import { useEffect, useState, useCallback } from "react";
import { BrowserRouter, Routes, Route, Link, useNavigate, useLocation } from "react-router-dom";
import { hospitalClient } from "./api/client";
import { ConnectError, Code } from "@connectrpc/connect";
import type { PatientDto, StaffMemberDto, EpisodeEventDto, RoomDto, SeverityReportEntry, DoctorStatEntry } from "./gen/proto/hospital/v1/hospital_pb";
import {
  Activity, Bed, Users, Stethoscope, AlertTriangle,
  PlayCircle, Bell, Search, Flame, Map, ArrowLeft, X, BarChart2
} from "lucide-react";

const Layout = ({ children }: { children: React.ReactNode }) => {
  const location = useLocation();
  return (
      <div className="flex h-screen bg-[#F3F4F6] text-slate-800 font-sans overflow-hidden">
        <aside className="w-64 bg-white border-r border-gray-200 flex flex-col justify-between shadow-sm z-10 shrink-0">
          <div>
            <div className="p-6 flex items-center gap-3 text-brand font-bold text-xl border-b border-gray-100">
              <Activity size={28} className="text-brand" />
              <div>
                <div className="leading-tight">Clínica Coste</div>
                <div className="text-xs text-gray-400 font-medium">Hospital ERP</div>
              </div>
            </div>
            <nav className="p-4 space-y-2">
              <Link to="/" className={`flex items-center gap-3 px-4 py-3 rounded-xl font-semibold transition ${location.pathname === '/' ? 'bg-blue-50 text-brand' : 'text-gray-500 hover:bg-gray-50'}`}>
                <Activity size={20} /> Dashboard
              </Link>
              <Link to="/map" className={`flex items-center gap-3 px-4 py-3 rounded-xl font-semibold transition ${location.pathname === '/map' ? 'bg-blue-50 text-brand' : 'text-gray-500 hover:bg-gray-50'}`}>
                <Map size={20} /> Mapa de Planta
              </Link>
              <Link to="/staff" className={`flex items-center gap-3 px-4 py-3 rounded-xl font-semibold transition ${location.pathname === '/staff' ? 'bg-blue-50 text-brand' : 'text-gray-500 hover:bg-gray-50'}`}>
                <Stethoscope size={20} /> Staff Médico
              </Link>
            </nav>
          </div>
        </aside>

        <main className="flex-1 overflow-y-auto flex flex-col">
          <header className="bg-white border-b border-gray-200 h-16 flex items-center justify-between px-8 sticky top-0 z-10 shrink-0">
            <div className="flex items-center bg-gray-100 px-4 py-2 rounded-lg w-96 text-gray-500">
              <Search size={18} className="mr-2" />
              <input type="text" placeholder="Buscar..." className="bg-transparent border-none outline-none w-full text-sm" />
            </div>
            <div className="flex items-center gap-6">
              <button className="text-gray-400 hover:text-brand relative"><Bell size={20} /></button>
              <div className="flex items-center gap-3 border-l border-gray-200 pl-6">
                <div className="text-right">
                  <div className="text-sm font-bold">Dra. Manuela Montoya</div>
                  <div className="text-xs text-gray-500">Neurología</div>
                </div>
                <div className="w-10 h-10 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center font-bold">MM</div>
              </div>
            </div>
          </header>
          <div className="flex-1 overflow-y-auto">{children}</div>
        </main>
      </div>
  );
};

const Dashboard = ({ patients, patientsInHallway, episodes, fullHistory, rooms, severityReport, newPatientName, setNewPatientName, newPatientLevel, setNewPatientLevel, handleAdmit, handleReportAttack, handleWake }: any) => {
  const navigate = useNavigate();

  const occupiedBeds = patients.filter((p: PatientDto) => p.assignedRoom > 0).length;
  const availableBeds = rooms.reduce((sum: number, room: RoomDto) => {
    const patientsInRoom = patients.filter((p: PatientDto) => p.assignedRoom === room.number).length;
    return sum + (room.capacity - patientsInRoom);
  }, 0);

  return (
      <div className="p-8 max-w-7xl mx-auto space-y-6">
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-2xl p-6 flex justify-between items-center border border-blue-100 shadow-sm">
          <div>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">Buen día, Dra. Manuela 👋</h1>
            <p className="text-gray-600 text-sm">Monitoreo en tiempo real activo.</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="bg-white p-5 rounded-2xl border border-gray-200 shadow-sm flex items-center justify-between">
            <div><p className="text-sm text-gray-500 font-medium">Pacientes Totales</p><h3 className="text-3xl font-bold">{patients.length}</h3></div>
            <div className="w-12 h-12 rounded-full bg-blue-50 flex items-center justify-center text-blue-600"><Users size={24}/></div>
          </div>
          <div className="bg-white p-5 rounded-2xl border border-gray-200 shadow-sm flex items-center justify-between">
            <div><p className="text-sm text-gray-500 font-medium">En Pasillo</p><h3 className="text-3xl font-bold text-orange-500">{patientsInHallway.length}</h3></div>
            <div className="w-12 h-12 rounded-full bg-orange-50 flex items-center justify-center text-orange-500"><AlertTriangle size={24}/></div>
          </div>
          <div className="bg-white p-5 rounded-2xl border border-gray-200 shadow-sm flex items-center justify-between">
            <div><p className="text-sm text-gray-500 font-medium">Camas Ocupadas</p><h3 className="text-3xl font-bold text-green-500">{occupiedBeds}</h3></div>
            <div className="w-12 h-12 rounded-full bg-green-50 flex items-center justify-center text-green-500"><Bed size={24}/></div>
          </div>
          <div className="bg-white p-5 rounded-2xl border border-gray-200 shadow-sm flex items-center justify-between">
            <div><p className="text-sm text-gray-500 font-medium">Total Eventos</p><h3 className="text-3xl font-bold text-purple-500">{fullHistory.length}</h3></div>
            <div className="w-12 h-12 rounded-full bg-purple-50 flex items-center justify-center text-purple-500"><Activity size={24}/></div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
              <h2 className="text-lg font-bold mb-4">Admitir Nuevo Paciente</h2>
              <form onSubmit={handleAdmit} className="flex gap-4 items-end">
                <div className="flex-1">
                  <label className="text-xs font-bold text-gray-500 mb-1 block">Nombre</label>
                  <input type="text" value={newPatientName} onChange={e => setNewPatientName(e.target.value)} className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-brand outline-none" placeholder="Ej. Yeimy" />
                </div>
                <div className="w-48">
                  <label className="text-xs font-bold text-gray-500 mb-1 block">Nivel</label>
                  <select value={newPatientLevel} onChange={e => setNewPatientLevel(e.target.value)} className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm bg-white outline-none">
                    <option value="Mild">Leve</option>
                    <option value="Moderate">Moderado</option>
                    <option value="Severe">Grave</option>
                  </select>
                </div>
                <button type="submit" className="bg-brand text-white px-6 py-2 rounded-lg text-sm font-bold shadow-sm">Admitir</button>
              </form>
            </div>

            <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
              <h2 className="text-lg font-bold mb-4">Pacientes en Pasillo (5.1)</h2>
              <div className="space-y-3 max-h-60 overflow-y-auto">
                {patientsInHallway.map((p: PatientDto) => (
                    <div key={p.id} className="flex items-center justify-between p-3 border border-gray-100 rounded-xl hover:bg-gray-50">
                      <div>
                        <div className="font-bold text-sm">{p.name}</div>
                        <div className="text-xs text-gray-500">{p.state === "Awake" ? "Despierto" : "Crisis (Dormido)"}</div>
                      </div>
                      <div className="flex gap-2">
                        {p.state === "Awake" ? (
                            <button onClick={() => handleReportAttack(p.id)} className="text-xs bg-orange-100 text-orange-600 px-3 py-1.5 rounded-lg font-bold"><PlayCircle size={14} className="inline mr-1"/> Ataque</button>
                        ) : (
                            <button onClick={() => handleWake(p.id)} className="text-xs bg-gray-100 text-gray-600 px-3 py-1.5 rounded-lg font-bold">Despertar</button>
                        )}
                      </div>
                    </div>
                ))}
                {patientsInHallway.length === 0 && <p className="text-sm text-gray-400 text-center py-4">No hay pacientes en el pasillo.</p>}
              </div>
            </div>

            <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold flex items-center gap-2"><BarChart2 size={18} className="text-brand"/> Reporte de Severidad (5.4)</h2>
                <span className="text-xs font-bold bg-red-50 text-red-600 px-2 py-1 rounded-md">Pacientes Graves</span>
              </div>
              <table className="w-full text-left border-collapse">
                <thead>
                <tr className="border-b border-gray-200 text-xs text-gray-500 uppercase">
                  <th className="py-2">Paciente</th>
                  <th className="py-2 text-right">Episodios</th>
                </tr>
                </thead>
                <tbody>
                {severityReport.length === 0 ? (
                    <tr><td colSpan={2} className="text-center py-4 text-sm text-gray-400">No hay reportes.</td></tr>
                ) : (
                    severityReport.map((r: SeverityReportEntry, i: number) => (
                        <tr key={i} className="border-b border-gray-100 last:border-0">
                          <td className="py-3 text-sm font-medium text-gray-700">{r.patientName}</td>
                          <td className="py-3 text-right">
                        <span className={`text-sm font-bold px-3 py-1 rounded-full ${r.episodeCount > 0 ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-500'}`}>
                          {r.episodeCount}
                        </span>
                          </td>
                        </tr>
                    ))
                )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="space-y-6">
            <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 relative overflow-hidden group cursor-pointer hover:border-blue-300 transition" onClick={() => navigate('/map')}>
              <h2 className="text-lg font-bold flex items-center gap-2 mb-4"><Flame size={18} className="text-orange-500"/> Estado de Habitaciones (5.3)</h2>
              <div className="space-y-4">
                <div className="flex justify-between items-center bg-gray-50 p-3 rounded-lg">
                  <span className="text-sm font-semibold text-gray-600">Camas Disponibles</span>
                  <span className={`text-lg font-bold ${availableBeds > 0 ? 'text-green-500' : 'text-red-500'}`}>{availableBeds}</span>
                </div>
                <div className="flex justify-between items-center bg-gray-50 p-3 rounded-lg">
                  <span className="text-sm font-semibold text-gray-600">Camas Ocupadas</span>
                  <span className="text-lg font-bold text-orange-500">{occupiedBeds}</span>
                </div>
              </div>
              <button className="w-full mt-6 bg-blue-50 text-brand py-2 rounded-lg text-sm font-bold group-hover:bg-brand group-hover:text-white transition flex items-center justify-center gap-2">
                Ver Mapa Interactivo <ArrowLeft size={16} className="rotate-180"/>
              </button>
            </div>

            <div className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 flex flex-col h-[500px]">
              <h2 className="text-lg font-bold mb-4">Eventos en Vivo</h2>
              <div className="flex-1 overflow-y-auto pr-2 space-y-4">
                {episodes.map((ep: EpisodeEventDto) => (
                    <div key={ep.id} className="relative pl-6 before:absolute before:left-2 before:top-2 before:bottom-0 before:w-0.5 before:bg-gray-100 last:before:hidden">
                      <div className="absolute left-0 top-1.5 w-4 h-4 rounded-full border-2 border-white bg-brand shadow-sm"></div>
                      <div className="text-xs font-bold text-gray-400 mb-1">{new Date(ep.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</div>
                      <div className="text-sm font-semibold text-gray-800">{ep.patientName}</div>
                      <div className="text-xs text-gray-500">Atención: {ep.attendingDoctor}</div>
                    </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
  );
};

const FloorMap = ({ rooms, patients, handleWake }: any) => {
  const navigate = useNavigate();
  const [selectedPatient, setSelectedPatient] = useState<PatientDto | null>(null);

  const getRoomColor = (occupant: PatientDto | undefined) => {
    if (!occupant) return "bg-slate-200 border-slate-300 text-slate-500 hover:bg-slate-300";
    if (occupant.narcolepsyLevel === "Severe") return "bg-red-500 border-red-700 text-white shadow-[0_0_20px_rgba(239,68,68,0.6)]";
    if (occupant.narcolepsyLevel === "Moderate") return "bg-orange-400 border-orange-600 text-white shadow-[0_0_15px_rgba(251,146,60,0.5)]";
    return "bg-green-500 border-green-700 text-white shadow-[0_0_15px_rgba(34,197,94,0.5)]";
  };

  return (
      <div className="p-8 h-full flex flex-col bg-[#E5E7EB] relative overflow-hidden">
        <button onClick={() => navigate('/')} className="absolute top-8 left-8 z-20 bg-white text-gray-700 px-4 py-2 rounded-xl shadow-md font-bold flex items-center gap-2 hover:bg-gray-50 transition border border-gray-200">
          <ArrowLeft size={18} /> Volver al Dashboard
        </button>

        <div className="text-center mb-8 relative z-10 mt-10">
          <h1 className="text-3xl font-black text-gray-800 tracking-tight">Mapa Isométrico - Piso 1</h1>
        </div>

        <div className="flex-1 flex items-center justify-center overflow-hidden">
          <div
              className="relative"
              style={{
                width: "800px",
                height: "500px",
                perspective: "1500px",
              }}
          >
            <div
                className="absolute top-1/2 left-1/2 flex flex-wrap gap-8 p-12 bg-white/60 backdrop-blur-md rounded-3xl border-4 border-white shadow-2xl transition-transform duration-700 justify-center"
                style={{
                  transform: "translate(-50%, -50%) rotateX(60deg) rotateZ(-45deg)",
                  transformStyle: "preserve-3d",
                  width: "max-content",
                  maxWidth: "1000px",
                }}
            >
              {rooms.map((room: RoomDto) => {
                const bedsInRoom = Array.from({ length: room.capacity }, (_, i) => i);
                return (
                    <div key={room.number} className="bg-white/50 p-4 rounded-2xl border-2 border-white shadow-sm flex flex-col gap-4 relative" style={{ transform: "translateZ(10px)" }}>

                      <div className="absolute -top-4 -left-4 bg-gray-800 text-white w-10 h-10 flex items-center justify-center rounded-lg font-black shadow-lg border-2 border-white z-20 text-sm">
                        {room.number}
                      </div>

                      {bedsInRoom.map((bedIndex) => {
                        const patientsInThisRoom = patients.filter((p: PatientDto) => p.assignedRoom === room.number);
                        const occupant = patientsInThisRoom[bedIndex];
                        return (
                            <div
                                key={bedIndex}
                                onClick={() => occupant && setSelectedPatient(occupant)}
                                className={`w-28 h-28 rounded-xl border-b-8 border-r-8 transition-all duration-300 relative group flex flex-col items-center justify-center ${getRoomColor(occupant)} ${occupant ? 'cursor-pointer hover:-translate-y-2' : 'cursor-default'}`}
                                style={{ transform: "translateZ(20px)" }}
                            >
                              {occupant ? (
                                  <>
                                    <Bed size={24} className="mt-1 animate-bounce" style={{ animationDuration: '3s' }}/>
                                    <span className="text-[10px] font-bold mt-2 uppercase tracking-wider text-center px-1 truncate w-full">{occupant.name}</span>
                                  </>
                              ) : (
                                  <div className="flex flex-col items-center opacity-50">
                                    <Bed size={20} className="mb-1" />
                                    <span className="text-[10px] font-bold uppercase">Libre</span>
                                  </div>
                              )}
                            </div>
                        );
                      })}
                    </div>
                );
              })}
            </div>
          </div>
        </div>

        {selectedPatient && (
            <div className="absolute inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm" onClick={() => setSelectedPatient(null)}>
              <div className="bg-white rounded-3xl shadow-2xl p-8 max-w-md w-full mx-4 relative animate-in fade-in zoom-in duration-200" onClick={e => e.stopPropagation()}>
                <button onClick={() => setSelectedPatient(null)} className="absolute top-4 right-4 p-2 bg-gray-100 rounded-full hover:bg-gray-200 transition text-gray-600"><X size={20} /></button>
                <div className="text-center">
                  <div className="w-20 h-20 bg-blue-50 rounded-full flex items-center justify-center mx-auto mb-4"><Bed size={40} className="text-brand" /></div>
                  <h2 className="text-2xl font-black text-gray-800 mb-1">Hab. {selectedPatient.assignedRoom}</h2>
                  <div className="bg-gray-50 rounded-2xl p-6 text-left space-y-4 mb-6 border border-gray-100">
                    <div><p className="text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">Nombre</p><p className="font-bold text-gray-800 text-lg">{selectedPatient.name}</p></div>
                  </div>
                  <button onClick={() => { handleWake(selectedPatient.id); setSelectedPatient(null); }} className="w-full bg-yellow-400 hover:bg-yellow-500 text-yellow-900 font-bold py-3 rounded-xl transition shadow-sm">☀️ Despertar Paciente</button>
                </div>
              </div>
            </div>
        )}
      </div>
  );
};

const StaffView = ({ doctorStats }: { doctorStats: DoctorStatEntry[] }) => {
  const navigate = useNavigate();

  return (
      <div className="p-8 max-w-7xl mx-auto space-y-6">
        <div className="flex justify-between items-end">
          <div>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">Historial por Doctor (5.2)</h1>
            <p className="text-gray-600 text-sm">Rendimiento del personal usando la función MyEpisodes() del backend.</p>
          </div>
          <button onClick={() => navigate('/')} className="bg-white border border-gray-200 text-gray-700 px-4 py-2 rounded-xl shadow-sm font-bold flex items-center gap-2 hover:bg-gray-50 transition">
            <ArrowLeft size={18} /> Volver
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {doctorStats.map((s: DoctorStatEntry) => (
              <div key={s.doctorId} className="bg-white rounded-2xl border border-gray-200 shadow-sm p-6 flex flex-col">
                <div className="flex items-center gap-4 mb-4">
                  <div className={`w-14 h-14 rounded-full flex items-center justify-center text-xl font-bold ${s.role === 'Doctor' ? 'bg-blue-100 text-blue-600' : 'bg-green-100 text-green-600'}`}>
                    {s.doctorName.charAt(0)}
                  </div>
                  <div>
                    <h3 className="text-lg font-bold text-gray-800">{s.doctorName}</h3>
                    <span className={`text-xs font-bold px-2 py-1 rounded-md ${s.role === 'Doctor' ? 'bg-blue-50 text-blue-600' : 'bg-green-50 text-green-600'}`}>
                  {s.role}
                </span>
                  </div>
                </div>

                <div className="mt-auto bg-gray-50 rounded-xl p-4 flex justify-between items-center">
                  <span className="text-sm font-semibold text-gray-600">Episodios Atendidos</span>
                  <span className="text-2xl font-bold text-gray-800">{s.episodeCount}</span>
                </div>
              </div>
          ))}
        </div>
      </div>
  );
};

function App() {
  const [patients, setPatients] = useState<PatientDto[]>([]);
  const [staff, setStaff] = useState<StaffMemberDto[]>([]);
  const [episodes, setEpisodes] = useState<EpisodeEventDto[]>([]);
  const [fullHistory, setFullHistory] = useState<EpisodeEventDto[]>([]);
  const [rooms, setRooms] = useState<RoomDto[]>([]);
  const [severityReport, setSeverityReport] = useState<SeverityReportEntry[]>([]);
  const [doctorStats, setDoctorStats] = useState<DoctorStatEntry[]>([]);

  const [newPatientName, setNewPatientName] = useState("");
  const [newPatientLevel, setNewPatientLevel] = useState("Severe");

  const loadData = useCallback(async () => {
    try {
      const [patientsRes, staffRes, historyRes, roomsRes, severityRes, doctorStatsRes] = await Promise.all([
        hospitalClient.getAllPatients({}),
        hospitalClient.getStaff({}),
        hospitalClient.getHistory({}),
        hospitalClient.getRooms({}),
        hospitalClient.getSeverityReport({}),
        hospitalClient.getDoctorStats({}),
      ]);
      setPatients(patientsRes.patients);
      setStaff(staffRes.staff);
      setEpisodes([...historyRes.episodes].reverse());
      setFullHistory(historyRes.episodes);
      setRooms(roomsRes.rooms);
      setSeverityReport(severityRes.entries);
      setDoctorStats(doctorStatsRes.entries);
    } catch (err) { console.error(err); }
  }, []);

  useEffect(() => {
    loadData();
    const controller = new AbortController();
    let timeoutId: NodeJS.Timeout | null = null;

    async function watchEpisodes() {
      try {
        for await (const ep of hospitalClient.streamEpisodes({}, { signal: controller.signal })) {
          setEpisodes((prev) => {
            if (prev.some(e => e.id === ep.id)) return prev;
            return [ep, ...prev].slice(0, 50);
          });

          if (!timeoutId) {
            timeoutId = setTimeout(() => { loadData(); timeoutId = null; }, 500);
          }
        }
      } catch (err) {
        if (err instanceof ConnectError && err.code !== Code.Canceled) console.error(err);
      }
    }
    watchEpisodes();
    return () => { controller.abort(); if (timeoutId) clearTimeout(timeoutId); };
  }, [loadData]);

  const handleAdmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPatientName) return;
    await hospitalClient.admitPatient({ name: newPatientName, age: 30, narcolepsyLevel: newPatientLevel });
    setNewPatientName("");
    loadData();
  };

  const handleReportAttack = async (patientId: string) => {
    await hospitalClient.reportSleepAttack({ patientId, location: "Pasillo" });
    loadData();
  };

  const handleWake = async (patientId: string) => {
    await hospitalClient.wakePatient({ patientId });
    loadData();
  };

  const patientsInHallway = patients.filter((p) => p.assignedRoom === 0);

  return (
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={
                               <Dashboard
                                  patients={patients} patientsInHallway={patientsInHallway} episodes={episodes}
                                  fullHistory={fullHistory} rooms={rooms} severityReport={severityReport}
                                  newPatientName={newPatientName} setNewPatientName={setNewPatientName}
                                  newPatientLevel={newPatientLevel} setNewPatientLevel={setNewPatientLevel}
                                  handleAdmit={handleAdmit} handleReportAttack={handleReportAttack}
                                  handleWake={handleWake}
                              />
                            } />
            <Route path="/map" element={<FloorMap rooms={rooms} patients={patients} handleWake={handleWake} />} />
            <Route path="/staff" element={<StaffView doctorStats={doctorStats} />} />
          </Routes>
        </Layout>
      </BrowserRouter>
  );
}

export default App;