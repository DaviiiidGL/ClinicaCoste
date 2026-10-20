// web/src/api/client.ts
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { HospitalService } from "../gen/proto/hospital/v1/hospital_pb";

// Usamos window.location.origin para que funcione con el proxy de Vite
const transport = createConnectTransport({
    baseUrl: window.location.origin,
});

export const hospitalClient = createClient(HospitalService, transport);