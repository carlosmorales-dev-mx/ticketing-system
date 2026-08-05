export type LogType = "success" | "warning" | "info";

export interface LogEntry {
  id: string;
  timestamp: string;
  message: string;
  type: LogType;
}

// Log de actividad en memoria (no persiste entre recargas a
// propósito: es una bitácora de "lo que pasó en esta sesión", como
// la consola de un terminal real). Cada acción real del usuario
// (reservar, pagar, cancelar, expirar, actualización por WebSocket)
// se refleja aquí.
export function useActivityLog() {
  const entries = useState<LogEntry[]>("activity-log", () => []);

  function log(message: string, type: LogType = "info") {
    const time = new Date().toLocaleTimeString("es-ES", { hour12: false });
    entries.value = [
      ...entries.value,
      { id: crypto.randomUUID(), timestamp: `[${time}]`, message, type },
    ].slice(-50); // no dejar crecer la bitácora indefinidamente
  }

  return { entries, log };
}
