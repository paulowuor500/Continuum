import type { Vault } from "../lib/api";

export default function BeneficiaryList({ vaults }: { vaults: Vault[] }) {
  const uniqueKeys = Array.from(new Set(vaults.map((vault) => vault.beneficiary_pubkey)));

  return (
    <div className="vault-list">
      {uniqueKeys.slice(0, 4).map((key, index) => (
        <div className="stat" key={key}>
          <span>Beneficiary {index + 1}</span>
          <b>{key.slice(0, 26)}...</b>
        </div>
      ))}
      {uniqueKeys.length === 0 ? <div className="empty">No beneficiaries loaded yet.</div> : null}
    </div>
  );
}
