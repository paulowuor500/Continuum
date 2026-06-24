import type { Vault } from "../lib/api";

export default function Timeline({ vault }: { vault?: Vault }) {
  const label = vault?.status === "DORMANT" ? "Recovery package released" : "Recovery package locked";

  return (
    <div className="timeline">
      {["Vault initialized", "Lightning proof accepted", "Inactivity threshold monitored", label].map((item, index) => (
        <div className="timeline-row" key={item}>
          <span className="dot" style={{ opacity: index === 3 && vault?.status !== "DORMANT" ? 0.35 : 1 }} />
          <div>
            <b>{item}</b>
            <p className="mono">
              {index === 0 ? "Encrypted package sealed client-side" : index === 1 ? "1 sat proof-of-life heartbeat" : index === 2 ? "Autonomous scheduler loop" : "Beneficiary-only reveal state"}
            </p>
          </div>
        </div>
      ))}
    </div>
  );
}
