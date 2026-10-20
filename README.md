# 🏥 Hospital de los Costeños con Narcolepsia

Sistema de gestión hospitalaria diseñado para administrar la caótica dinámica de una clínica costeña, donde los pacientes sufren de narcolepsia severa y pueden quedarse dormidos en cualquier momento.

El proyecto está construido con un backend concurrente en **Go** (utilizando ConnectRPC) y un frontend interactivo en **React + TypeScript**.

## 👤 Autor

- David Giraldo Lema

---

## 🛠️ Tech Stack

### Backend

- **Go (1.21+)**: Lógica de dominio, concurrencia (`sync.RWMutex`, goroutines) y colas de prioridad (`container/heap`).
- **ConnectRPC (v2)** y **Protocol Buffers**: API tipada con soporte para Server-Streaming.
- **Buf CLI**: Generación automática de código (Go y TypeScript) a partir del `.proto`.

### Frontend

- **React (18+)** y **TypeScript**: Interfaz de usuario basada en componentes.
- **Vite**: Empaquetador ultrarrápido.
- **Tailwind CSS**: Estilos del dashboard.
- **Connect-ES**: Cliente generado para consumir los RPCs.

---

## 🚀 Instrucciones de Instalación

Sigue estos pasos para ejecutar el proyecto desde cero en tu máquina local.

### Prerrequisitos

- [Go (Golang)](https://go.dev/dl/) **1.21** o superior.
- [Node.js](https://nodejs.org/) **18.x** o superior.
- [Buf CLI](https://buf.build/docs/installation) (opcional, solo si deseas modificar el `.proto`).

### 1. Clonar el repositorio

```bash
git clone https://github.com/DaviiiidGL/ClinicaCoste.git
cd ClinicaCoste
```

### 2. Instalar dependencias

```bash
# Backend
go mod download

# Frontend
cd web
npm install
cd ..
```

### 3. Ejecutar la interfaz gráfica (Frontend + API)

El proyecto usa un servidor HTTP con ConnectRPC y un frontend en React. Abre **dos terminales** en la raíz del proyecto.

**Terminal 1 (Backend - Interfaz):**

```bash
go run .
```

El servidor web estará escuchando en `http://localhost:8080`.

**Terminal 2 (Frontend - React):**

```bash
cd web
npm run dev
```

Abre tu navegador en `http://localhost:5173`.

### 4. Ejecutar el escenario por consola (Punto 6 del PDF)

Si deseas ver la demostración por consola que exige el profesor (creación del hospital, ataques, asignación de camas, etc.), ejecuta:

```bash
go run ./demo
```

### 5. Correr los tests

Para verificar que el dominio funciona y no existen condiciones de carrera (*data races*):

```bash
go test -race ./...
```

---

## 📖 Instrucciones de Uso

El sistema tiene dos modos de uso: la interfaz gráfica (para monitoreo en vivo) y el modo consola (para validar el escenario).

### A. Interfaz gráfica (web)

Abre `http://localhost:5173` en tu navegador. Usa el menú lateral para navegar:

1. **Dashboard:**
   - Usa el formulario para **Admitir Paciente** (nombre y gravedad).
   - Revisa la lista de **Pacientes en Pasillo** (Consulta 5.1) y el **Reporte de Severidad** (Consulta 5.4).
   - Observa el **Log de Eventos en Vivo**, que se actualiza en tiempo real gracias a ConnectRPC.
2. **Mapa de Planta:** Haz clic en el menú "Mapa de Planta" o en la tarjeta de habitaciones.
   - Verás un **mapa isométrico 3D** con las camas ocupadas coloreadas según la gravedad del paciente (*heatmap*).
   - Haz clic en una cama ocupada para **despertar** al paciente.
3. **Staff Médico:** Consulta el rendimiento del personal (Consulta 5.2). Verás cuántos episodios ha atendido cada doctor, enfermero o camillero.

### B. Modo consola (escenario de evaluación)

Si ejecutaste `go run ./demo`, la terminal imprimirá paso a paso el escenario exacto que pide el PDF:

- Creación del hospital, contratación de doctores y asignación de camas.
- Ataques de sueño con asignación exitosa y el caso de error "No room available".
- Despertar de un paciente y reasignación automática desde el pasillo.
- La impresión final de las cuatro consultas de la Sección 5.

---

## ✨ Características principales

- **Diseño orientado a objetos en Go:** Uso de *Composition over Inheritance* (embedding de `Person`) y polimorfismo con la interfaz `Attender`.
- **Simulación concurrente:** Goroutines que hacen que los pacientes sufran ataques de sueño de forma aleatoria, protegidos por `sync.RWMutex`.
- **Autoasignación:** Al liberar una cama, el sistema busca en la cola de prioridad (`container/heap`) al paciente más grave en el pasillo y lo ingresa automáticamente.
- **API ConnectRPC:** Comunicación moderna y tipada con soporte para Server-Streaming (WebSockets/HTTP2).
- **Frontend interactivo:** Dashboard en React con mapa isométrico 3D, reportes en tiempo real y métricas del personal.
