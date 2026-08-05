// En un sistema real esto vendría de un login. Para esta demo,
// generamos un UUID por navegador y lo persistimos en localStorage,
// así el mismo usuario mantiene su identidad entre recargas.
export function useUserId(): string {
  if (import.meta.server) {
    // En el servidor (SSR) no hay localStorage; se resuelve en el
    // cliente cuando el componente se monta.
    return "";
  }

  const STORAGE_KEY = "ticketing_user_id";
  let id = localStorage.getItem(STORAGE_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(STORAGE_KEY, id);
  }
  return id;
}
