import { createContext, useContext } from "react";
import type { Choice } from "../types";

// Админку видно, если в localStorage стоит is_admin = true. Генерировать и сохранять
// сервер всё равно даёт только при запуске с -admin — флаг в браузере ставит кто угодно.
export const adminRequested = () => localStorage.getItem("is_admin") === "true";

export type Admin = {
  enabled: boolean;
  active: string | null; // id открытой записи, «w:дом»
  open(id: string): void;
  choices: Record<string, Choice>;
};

export const AdminContext = createContext<Admin>({ enabled: false, active: null, open() {}, choices: {} });

export const useAdmin = () => useContext(AdminContext);
