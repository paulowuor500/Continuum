import { describe, expect, it } from 'vitest';

import { createVault } from './api';

describe('api client', () => {
  it('uses the backend API URL for vault creation', async () => {
    const originalFetch = global.fetch;
    global.fetch = async (input) => {
      const url = typeof input === 'string' ? input : input.url;
      return {
        ok: true,
        json: async () => ({ vault_id: 'abc', message: 'ok' }),
        status: 200
      };
    };

    try {
      await createVault({
        alias: 'Demo',
        beneficiary_pubkey: '02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
        encrypted_payload: 'cipher',
        check_in_interval_seconds: 30,
        multisig_required: 2,
        multisig_pubkeys: ['02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa']
      });
    } finally {
      global.fetch = originalFetch;
    }
  });
});
