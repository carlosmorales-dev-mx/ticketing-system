// Cuántos asientos puede apartar una misma persona a la vez en este
// proyecto. El backend lo valida igual (reservation.MaxPendingPerUser,
// error MAX_SEATS_PER_USER); aquí solo sirve para no dejarte intentar.
export const MAX_SEATS_PER_USER = 3;

// "3C" -> "Fila 3, asiento C". Si el formato no coincide, devuelve el
// texto tal cual para no romper nada.
export function describeSeat(seatLabel: string): string {
  const match = /^(\d+)\s*(.+)$/.exec(seatLabel.trim());
  return match ? `Fila ${match[1]}, asiento ${match[2]}` : seatLabel;
}
