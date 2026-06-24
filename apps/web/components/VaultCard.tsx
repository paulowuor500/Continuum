import type { Vault } from "../lib/api";
import StatusBadge from "./StatusBadge";

interface VaultCardProps {
  vault: Vault;
  selected: boolean;
  onSelect: (vault: Vault) => void;
  onCheckIn: (vault: Vault) => void;
  onWarp: (vault: Vault) => void;
  busy: boolean;
}

function timeAgo(value: string) {
  const then = new Date(value).getTime();
  const diff = Math.max(0, Date.now() - then);
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function intervalLabel(seconds: number) {
  if (seconds < 120) return `${seconds}s`;
  if (seconds < 7200) return `${Math.round(seconds / 60)}m`;
  if (seconds < 172800) return `${Math.round(seconds / 3600)}h`;
  return `${Math.round(seconds / 86400)}d`;
}

export default function VaultCard({ vault, selected, onSelect, onCheckIn, onWarp, busy }: VaultCardProps) {
  return (
    <article className="vault-card" style={{ borderColor: selected ? "var(--cyan)" : undefined }}>
      <div className="vault-top">
        <div className="vault-title">
          <h3>{vault.alias}</h3>
          <span className="mono truncate">{vault.id}</span>
        </div>
        <StatusBadge status={vault.status} />
      </div>

      <div className="vault-stats">
        <div className="stat">
          <span>Last proof</span>
          <b>{timeAgo(vault.last_check_in_at)}</b>
        </div>
        <div className="stat">
          <span>Window</span>
          <b>{intervalLabel(vault.check_in_interval_seconds)}</b>
        </div>
        <div className="stat">
          <span>Beneficiary key</span>
          <b>{vault.beneficiary_pubkey.slice(0, 18)}...</b>
        </div>
        <div className="stat">
          <span>Multisig</span>
          <b>
            {vault.multisig_required}-of-{vault.multisig_pubkeys.length}
          </b>
        </div>
        <div className="stat">
          <span>Funding address</span>
          <b>{vault.multisig_address || "RPC preview"}</b>
        </div>
      </div>

      <div className="hero-actions">
        <button className="button" type="button" onClick={() => onSelect(vault)}>
          Inspect
        </button>
        <button className="button primary" type="button" onClick={() => onCheckIn(vault)} disabled={busy}>
          Check in
        </button>
        <button className="button danger" type="button" onClick={() => onWarp(vault)} disabled={busy}>
          Time warp
        </button>
      </div>
    </article>
  );
}
