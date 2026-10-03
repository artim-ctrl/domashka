import { HomePage } from "./pages/Home";
import { RecordPage } from "./pages/Record";

// Страниц две, так что без роутера: переходы — обычные ссылки.
export function App() {
  return location.pathname.replace(/\/+$/, "") === "/record" ? <RecordPage /> : <HomePage />;
}
