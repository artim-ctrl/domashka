import { createContext, useContext } from "react";
import type { Choices } from "../api/types";
import type { ClipId } from "../lib/clip";

export const adminRequested = () => localStorage.getItem("is_admin") === "true";

export type Admin = {
    enabled: boolean;
    active: ClipId | null;
    open(id: ClipId): void;
    choices: Choices;
};

export const AdminContext = createContext<Admin>({ enabled: false, active: null, open: () => {}, choices: {} });

export const useAdmin = () => useContext(AdminContext);
