export type VaultStatus = "ACTIVE" | "DORMANT";

export interface Vault {
  id: string;
  alias: string;
  beneficiary_pubkey: string;
  encrypted_payload: string;
  check_in_interval_seconds: number;
  last_check_in_at: string;
  status: VaultStatus;
  multisig_required: number;
  multisig_pubkeys: string[];
  multisig_address: string;
  multisig_redeem_script: string;
  multisig_descriptor: string;
  multisig_network: string;
}

export interface VaultStatusResponse {
  id: string;
  alias: string;
  status: VaultStatus;
  last_seen: string;
  payload_locked: boolean;
  message?: string;
  encrypted_payload?: string;
  multisig_required: number;
  multisig_pubkeys: string[];
  multisig_address: string;
  multisig_redeem_script?: string;
  multisig_descriptor: string;
  multisig_network: string;
}

function apiBaseUrl() {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  if (typeof window !== "undefined") {
    if (window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1") {
      return "http://localhost:8080/api";
    }
    return `${window.location.origin}/api`;
  }
  return "http://localhost:8080/api";
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl()}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    cache: "no-store"
  });

  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(body.error ?? `Continuum API returned ${response.status}`);
  }
  return body as T;
}

export async function listVaults(): Promise<Vault[]> {
  const data = await request<{ vaults: Vault[] }>("/vaults");
  return data.vaults;
}

export async function createVault(input: {
  alias: string;
  beneficiary_pubkey: string;
  encrypted_payload: string;
  check_in_interval_seconds: number;
  multisig_required: number;
  multisig_pubkeys: string[];
}): Promise<{ vault_id: string; message: string }> {
  return request("/vaults", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export async function getVaultStatus(id: string): Promise<VaultStatusResponse> {
  return request(`/vaults/${id}`);
}

export async function checkInVault(id: string): Promise<{ message: string }> {
  return request(`/vaults/${id}/check-in`, { method: "POST" });
}

export async function updateVaultTimer(id: string, seconds: number): Promise<{ message: string }> {
  return request(`/vaults/${id}/timer`, {
    method: "PATCH",
    body: JSON.stringify({ check_in_interval_seconds: seconds })
  });
}

export async function addBeneficiary(id: string, pubkey: string): Promise<{ message: string }> {
  return request(`/vaults/${id}/beneficiaries`, {
    method: "POST",
    body: JSON.stringify({ pubkey })
  });
}

export async function warpVault(id: string): Promise<{ message: string }> {
  return request(`/vaults/${id}/warp`, { method: "POST" });
}

export async function createInvoice(id: string): Promise<{ invoice: string }> {
  return request(`/vaults/${id}/invoice`, { method: "POST" });
}
