import pg from 'pg';

const { Client } = pg;

const connectionString = process.env.DATABASE_URL || 'postgres://postgres:postgres@localhost:5432/continuum?sslmode=disable';

async function seedDatabase() {
  const client = new Client({ connectionString });
  
  console.log("🌱 Connecting to Continuum relational engine for seed initialization...");
  await client.connect();

  try {
    // 1. Clear out stale structural debris safely
    await client.query('TRUNCATE TABLE vaults RESTART IDENTITY CASCADE;');

    // Mock Payload Capsule Data sets (Base64 data packets mimicking browser encryption outputs)
    const activePayload = Buffer.from(JSON.stringify({
      ephemeralPubKey: "03e9a26312...",
      iv: "af3e99...",
      authTag: "99bbcf...",
      ciphertext: "77a8b9cedf"
    })).toString('base64');

    const dormantPayload = Buffer.from(JSON.stringify({
      ephemeralPubKey: "02b41178cd...",
      iv: "bc41de...",
      authTag: "12faaa...",
      ciphertext: "883109bcfa"
    })).toString('base64');

    console.log("📥 Injecting hackathon user demo records...");
    const ownerPubkey = '02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
    const operatorPubkey = '03bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb';
    const activeBeneficiary = '02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d';
    const dormantBeneficiary = '03b41178cd224b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1df27b';

    // Record A: Active Shielded State Scenario
    await client.query(`
      INSERT INTO vaults (id, alias, beneficiary_pubkey, encrypted_payload, check_in_interval_seconds, last_check_in_at, status, created_at, updated_at)
      VALUES (
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'Satoshi_Legacy_Vault',
        $2,
        $1,
        86400, -- 24 Hour check-in safety window
        NOW(), -- Checked in just now
        'ACTIVE',
        NOW(),
        NOW()
      );
    `, [activePayload, activeBeneficiary, ownerPubkey, operatorPubkey]);

    // Record B: Expired Dormant Stage Scenario (Ready for Bob to extract on stage)
    await client.query(`
      INSERT INTO vaults (id, alias, beneficiary_pubkey, encrypted_payload, check_in_interval_seconds, last_check_in_at, status, created_at, updated_at)
      VALUES (
        'b11ebc99-9c0b-4ef8-bb6d-6bb9bd380b22',
        'Lost_Operator_Alpha',
        $2,
        $1,
        30, -- Demo fast timeout window
        NOW() - INTERVAL '5 days', -- Last seen 5 days ago (Guarantees instant timeout status)
        'ACTIVE', -- The background scheduler.go file will instantly flip this to DORMANT on its first loop
        NOW() - INTERVAL '5 days',
        NOW() - INTERVAL '5 days'
      );
    `, [dormantPayload, dormantBeneficiary, ownerPubkey, operatorPubkey]);

    console.log("==============================================================================");
    console.log("🎉 SUCCESS: Database seeding parameters completed flawlessly!");
    console.log("🛡️ Active Vault ID: a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11");
    console.log("⚠️ Dormant Target ID: b11ebc99-9c0b-4ef8-bb6d-6bb9bd380b22");
    console.log("==============================================================================");

  } catch (error) {
    console.error("❌ Seeding execution failure: ", error);
  } finally {
    await client.end();
  }
}

seedDatabase();
