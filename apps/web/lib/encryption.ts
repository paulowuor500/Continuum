interface EncryptedPackage {
  version: 1;
  keyHint: string;
  iv: string;
  ciphertext: string;
}

const encoder = new TextEncoder();
const decoder = new TextDecoder();

function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}

function base64ToBytes(value: string): Uint8Array {
  const binary = atob(value);
  return Uint8Array.from(binary, (char) => char.charCodeAt(0));
}

function asArrayBuffer(bytes: Uint8Array): ArrayBuffer {
  return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength) as ArrayBuffer;
}

async function deriveAesKey(keyMaterial: string): Promise<CryptoKey> {
  const digest = await crypto.subtle.digest("SHA-256", encoder.encode(keyMaterial.trim()));
  return crypto.subtle.importKey("raw", digest, "AES-GCM", false, ["encrypt", "decrypt"]);
}

export async function encryptPayload(plaintext: string, beneficiaryKeyMaterial: string): Promise<string> {
  try {
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const aesKey = await deriveAesKey(beneficiaryKeyMaterial);
    const encrypted = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, aesKey, encoder.encode(plaintext));

    const keyHash = await crypto.subtle.digest("SHA-256", encoder.encode(beneficiaryKeyMaterial));
    const keyHint = bytesToBase64(new Uint8Array(keyHash).slice(0, 6));
    const payload: EncryptedPackage = {
      version: 1,
      keyHint,
      iv: bytesToBase64(iv),
      ciphertext: bytesToBase64(new Uint8Array(encrypted))
    };

    return btoa(JSON.stringify(payload));
  } catch (error) {
    throw new Error(`Browser encryption failed: ${(error as Error).message}`);
  }
}

export async function decryptPayload(base64Package: string, beneficiaryKeyMaterial: string): Promise<string> {
  try {
    const payload = JSON.parse(atob(base64Package)) as EncryptedPackage;
    const aesKey = await deriveAesKey(beneficiaryKeyMaterial);
    const plaintext = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: asArrayBuffer(base64ToBytes(payload.iv)) },
      aesKey,
      asArrayBuffer(base64ToBytes(payload.ciphertext))
    );

    return decoder.decode(plaintext);
  } catch {
    throw new Error("Local decryption failed: invalid encrypted package or beneficiary key material.");
  }
}
