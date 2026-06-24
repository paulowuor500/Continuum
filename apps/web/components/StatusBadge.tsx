import type { VaultStatus } from "../lib/api";

export default function StatusBadge({ status }: { status: VaultStatus }) {
  return <span className={`status ${status === "ACTIVE" ? "active" : "dormant"}`}>{status}</span>;
}
