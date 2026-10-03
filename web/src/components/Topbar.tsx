import type { ReactNode } from "react";
import { LangSwitch } from "./LangSwitch";
import { ThemeSwitch } from "./ThemeSwitch";

export function Topbar({ children }: { children?: ReactNode }) {
  return (
    <header className="topbar">
      <div className="topbar__main">{children}</div>
      <LangSwitch />
      <ThemeSwitch />
    </header>
  );
}
