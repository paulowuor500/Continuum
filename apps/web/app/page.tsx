"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  addBeneficiary,
  checkInVault,
  createInvoice,
  createVault,
  getVaultStatus,
  listVaults,
  updateVaultTimer,
  type Vault,
  type VaultStatusResponse
} from "../lib/api";
import { compactInvoice, demoInvoice } from "../lib/lightning";

const ownerPubkey = "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
const operatorPubkey = "03bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";

const demoVaults: Vault[] = [
  {
    id: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
    alias: "Family Cold Storage",
    beneficiary_pubkey: "02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d",
    encrypted_payload: "eyJjaXBoZXJ0ZXh0IjoiZGVtby1zaGllbGRlZCJ9",
    check_in_interval_seconds: 86400,
    last_check_in_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    status: "ACTIVE",
    multisig_required: 2,
    multisig_pubkeys: [
      "02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d",
      ownerPubkey,
      operatorPubkey
    ],
    multisig_address: "offline-regtest-preview",
    multisig_redeem_script: "",
    multisig_descriptor:
      "wsh(sortedmulti(2,02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d,03bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb))",
    multisig_network: "regtest"
  }
];

function formatDuration(seconds: number) {
  if (seconds < 120) return `${seconds} seconds`;
  if (seconds < 7200) return `${Math.round(seconds / 60)} minutes`;
  if (seconds < 172800) return `${Math.round(seconds / 3600)} hours`;
  return `${Math.round(seconds / 86400)} days`;
}

function timeAgo(value: string) {
  const diff = Math.max(0, Date.now() - new Date(value).getTime());
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function payloadForDemo(alias: string, instructions: string) {
  return window.btoa(JSON.stringify({ alias, ciphertext: instructions }));
}

function isNetworkFallback(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  return /fetch|network|failed to fetch|load failed|connection|refused/i.test(message);
}

export default function Home() {
  const [vaults, setVaults] = useState<Vault[]>(demoVaults);
  const [selectedId, setSelectedId] = useState(demoVaults[0].id);
  const [status, setStatus] = useState<VaultStatusResponse | null>(null);
  const [invoice, setInvoice] = useState(demoInvoice(demoVaults[0].id));
  const [timerDraft, setTimerDraft] = useState(String(demoVaults[0].check_in_interval_seconds));
  const [beneficiaryDraft, setBeneficiaryDraft] = useState("");
  const [message, setMessage] = useState("Demo mode ready. Pay 1 sat to prove life.");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [apiOnline, setApiOnline] = useState(false);

  const selected = useMemo(() => vaults.find((vault) => vault.id === selectedId) ?? vaults[0], [selectedId, vaults]);
  const beneficiaries = selected?.multisig_pubkeys ?? [];

  function localStatusFor(vault: Vault): VaultStatusResponse {
    return {
      id: vault.id,
      alias: vault.alias,
      status: vault.status,
      last_seen: vault.last_check_in_at,
      payload_locked: vault.status === "ACTIVE",
      encrypted_payload: vault.status === "DORMANT" ? vault.encrypted_payload : undefined,
      message: vault.status === "ACTIVE" ? "Recovery package locked while proof-of-life is current." : undefined,
      multisig_required: vault.multisig_required,
      multisig_pubkeys: vault.multisig_pubkeys,
      multisig_address: vault.multisig_address,
      multisig_redeem_script: vault.multisig_redeem_script,
      multisig_descriptor: vault.multisig_descriptor,
      multisig_network: vault.multisig_network
    };
  }

  async function refresh(silent = false) {
    try {
      const liveVaults = await listVaults();
      setVaults(liveVaults.length > 0 ? liveVaults : demoVaults);
      setApiOnline(true);
      setError("");
      if (!silent) setMessage("Live API state loaded.");
    } catch (err) {
      setApiOnline(false);
      setVaults((current) => (current.length > 0 ? current : demoVaults));
      setError(isNetworkFallback(err) ? "" : (err as Error).message);
      if (!silent) setMessage("Offline demo mode. Actions will update this screen locally.");
    }
  }

  useEffect(() => {
    refresh(true);
    const timer = window.setInterval(() => refresh(true), 10000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!selected) return;
    setTimerDraft(String(selected.check_in_interval_seconds));

    getVaultStatus(selected.id)
      .then((data) => {
        setStatus(data);
        setError("");
      })
      .catch(() => setStatus(localStatusFor(selected)));

    createInvoice(selected.id)
      .then((data) => setInvoice(data.invoice))
      .catch(() => setInvoice(demoInvoice(selected.id)));
  }, [selected]);

  function updateSelectedVault(mutator: (vault: Vault) => Vault) {
    setVaults((current) => current.map((vault) => (vault.id === selected.id ? mutator(vault) : vault)));
  }

  async function payProofOfLife() {
    if (!selected) return;
    setBusy("pay");
    setError("");
    try {
      const result = await checkInVault(selected.id);
      setMessage(result.message || "1 sat proof-of-life accepted.");
      await refresh(true);
    } catch (err) {
      updateSelectedVault((vault) => ({ ...vault, status: "ACTIVE", last_check_in_at: new Date().toISOString() }));
      setApiOnline(false);
      setMessage("1 sat proof-of-life accepted in demo mode.");
      setError(isNetworkFallback(err) ? "" : (err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function saveTimer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selected) return;
    const seconds = Number(timerDraft);
    if (!Number.isFinite(seconds) || seconds < 30) {
      setError("Timer must be at least 30 seconds.");
      return;
    }

    setBusy("timer");
    setError("");
    try {
      const result = await updateVaultTimer(selected.id, seconds);
      setMessage(result.message || "Inactivity timer updated.");
      await refresh(true);
    } catch (err) {
      updateSelectedVault((vault) => ({ ...vault, check_in_interval_seconds: seconds }));
      setApiOnline(false);
      setMessage(`Timer updated locally to ${formatDuration(seconds)}.`);
      setError(isNetworkFallback(err) ? "" : (err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function addBeneficiarySigner(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selected) return;
    const pubkey = beneficiaryDraft.trim();
    if (!/^(02|03)[0-9a-fA-F]{64}$/.test(pubkey)) {
      setError("Enter a compressed secp256k1 public key starting with 02 or 03.");
      return;
    }

    setBusy("beneficiary");
    setError("");
    try {
      const result = await addBeneficiary(selected.id, pubkey);
      setMessage(result.message || "Beneficiary added.");
      setBeneficiaryDraft("");
      await refresh(true);
    } catch (err) {
      updateSelectedVault((vault) => {
        const pubkeys = Array.from(new Set([...vault.multisig_pubkeys, pubkey.toLowerCase()])).sort();
        return {
          ...vault,
          multisig_pubkeys: pubkeys,
          multisig_descriptor: `wsh(sortedmulti(${vault.multisig_required},${pubkeys.join(",")}))`,
          multisig_address: "offline-regtest-preview"
        };
      });
      setBeneficiaryDraft("");
      setApiOnline(false);
      setMessage("Beneficiary signer added locally.");
      setError(isNetworkFallback(err) ? "" : (err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function createDemoVault(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const alias = String(form.get("alias") || "New Recovery Vault");
    const beneficiary = String(form.get("beneficiary") || demoVaults[0].beneficiary_pubkey);
    const seconds = Number(form.get("interval") || 86400);
    const instructions = String(form.get("instructions") || "Recovery instructions sealed for beneficiary.");
    const pubkeys = [beneficiary, ownerPubkey, operatorPubkey];

    setBusy("create");
    try {
      const result = await createVault({
        alias,
        beneficiary_pubkey: beneficiary,
        encrypted_payload: payloadForDemo(alias, instructions),
        check_in_interval_seconds: seconds,
        multisig_required: 2,
        multisig_pubkeys: pubkeys
      });
      setSelectedId(result.vault_id);
      setMessage("Vault created.");
      await refresh(true);
      event.currentTarget.reset();
    } catch (err) {
      const localVault: Vault = {
        id: crypto.randomUUID(),
        alias,
        beneficiary_pubkey: beneficiary,
        encrypted_payload: payloadForDemo(alias, instructions),
        check_in_interval_seconds: seconds,
        last_check_in_at: new Date().toISOString(),
        status: "ACTIVE",
        multisig_required: 2,
        multisig_pubkeys: pubkeys.sort(),
        multisig_address: "offline-regtest-preview",
        multisig_redeem_script: "",
        multisig_descriptor: `wsh(sortedmulti(2,${pubkeys.sort().join(",")}))`,
        multisig_network: "regtest"
      };
      setVaults((current) => [localVault, ...current]);
      setSelectedId(localVault.id);
      setMessage("Vault created locally.");
      setError(isNetworkFallback(err) ? "" : (err as Error).message);
    } finally {
      setBusy("");
    }
  }

  return (
    <main className="console-shell">
      <header className="console-header">
        <div>
          <p className="eyebrow">Continuum</p>
          <h1>Proof-of-life recovery console</h1>
        </div>
        <div className={`mode-pill ${apiOnline ? "live" : ""}`}>{apiOnline ? "Live API" : "Demo mode"}</div>
      </header>

      <section className="workspace">
        <aside className="rail">
          <div className="panel compact">
            <label className="field">
              <span>Vault</span>
              <select value={selectedId} onChange={(event) => setSelectedId(event.target.value)}>
                {vaults.map((vault) => (
                  <option key={vault.id} value={vault.id}>
                    {vault.alias}
                  </option>
                ))}
              </select>
            </label>
            <div className="status-block">
              <span className={`status ${selected?.status === "ACTIVE" ? "active" : "dormant"}`}>{selected?.status}</span>
              <b>{selected ? timeAgo(selected.last_check_in_at) : "n/a"}</b>
              <small>Last proof</small>
            </div>
          </div>

          <form className="panel compact" onSubmit={createDemoVault}>
            <div className="panel-title">
              <p className="eyebrow">New vault</p>
              <h2>Create</h2>
            </div>
            <label className="field">
              <span>Alias</span>
              <input name="alias" placeholder="Family treasury" />
            </label>
            <label className="field">
              <span>Beneficiary key</span>
              <input name="beneficiary" defaultValue={demoVaults[0].beneficiary_pubkey} />
            </label>
            <label className="field">
              <span>Timer</span>
              <select name="interval" defaultValue="86400">
                <option value="30">30 seconds</option>
                <option value="3600">1 hour</option>
                <option value="86400">24 hours</option>
                <option value="2592000">30 days</option>
              </select>
            </label>
            <label className="field">
              <span>Recovery note</span>
              <textarea name="instructions" placeholder="Encrypted handoff instructions" />
            </label>
            <button className="button primary" disabled={busy === "create"} type="submit">
              Create vault
            </button>
          </form>
        </aside>

        <section className="main-stack">
          {message ? <div className="notice">{message}</div> : null}
          {error ? <div className="notice error">{error}</div> : null}

          <section className="proof-panel">
            <div>
              <p className="eyebrow">Proof-of-life</p>
              <h2>Pay 1 sat to keep recovery locked</h2>
              <p className="muted">
                The Lightning payment records activity and resets the vault to active. In demo mode, the same action
                updates the local vault state.
              </p>
            </div>
            <div className="invoice-box">
              <div className="qr-grid" aria-label="Stylized Lightning QR" />
              <code>{compactInvoice(invoice)}</code>
              <button className="button primary" disabled={busy === "pay"} onClick={payProofOfLife} type="button">
                Pay 1 sat
              </button>
            </div>
          </section>

          <section className="control-grid">
            <form className="panel" onSubmit={saveTimer}>
              <div className="panel-title">
                <p className="eyebrow">Inactivity timer</p>
                <h2>{selected ? formatDuration(selected.check_in_interval_seconds) : "n/a"}</h2>
              </div>
              <label className="field">
                <span>Trigger after</span>
                <select value={timerDraft} onChange={(event) => setTimerDraft(event.target.value)}>
                  <option value="30">30 seconds</option>
                  <option value="300">5 minutes</option>
                  <option value="3600">1 hour</option>
                  <option value="86400">24 hours</option>
                  <option value="604800">7 days</option>
                  <option value="2592000">30 days</option>
                </select>
              </label>
              <button className="button" disabled={busy === "timer"} type="submit">
                Save timer
              </button>
            </form>

            <form className="panel" onSubmit={addBeneficiarySigner}>
              <div className="panel-title">
                <p className="eyebrow">Beneficiaries</p>
                <h2>
                  {selected?.multisig_required}-of-{beneficiaries.length}
                </h2>
              </div>
              <div className="beneficiary-list">
                {beneficiaries.map((key, index) => (
                  <div className="beneficiary-row" key={key}>
                    <span>{index === 0 ? "Primary" : `Signer ${index + 1}`}</span>
                    <code>{key.slice(0, 18)}...{key.slice(-8)}</code>
                  </div>
                ))}
              </div>
              <label className="field">
                <span>Add beneficiary public key</span>
                <input value={beneficiaryDraft} onChange={(event) => setBeneficiaryDraft(event.target.value)} placeholder="02..." />
              </label>
              <button className="button" disabled={busy === "beneficiary"} type="submit">
                Add beneficiary
              </button>
            </form>
          </section>

          <section className="panel policy-panel">
            <div className="panel-title">
              <p className="eyebrow">Recovery policy</p>
              <h2>{status?.payload_locked ? "Payload locked" : "Payload released"}</h2>
            </div>
            <div className="detail-grid">
              <div>
                <span>Funding address</span>
                <code>{status?.multisig_address || "Start regtest API for address"}</code>
              </div>
              <div>
                <span>Descriptor</span>
                <code>{status?.multisig_descriptor}</code>
              </div>
              <div>
                <span>Recovery package</span>
                <code>{status?.encrypted_payload ?? status?.message ?? "No package available"}</code>
              </div>
            </div>
          </section>
        </section>
      </section>
    </main>
  );
}
