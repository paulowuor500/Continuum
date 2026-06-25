CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS vaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alias VARCHAR(255) NOT NULL,
    beneficiary_pubkey TEXT NOT NULL,
    encrypted_payload TEXT NOT NULL,
    check_in_interval_seconds INT NOT NULL,
    last_check_in_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    multisig_required INT NOT NULL DEFAULT 2,
    multisig_pubkeys TEXT[] NOT NULL DEFAULT '{}',
    multisig_address TEXT NOT NULL DEFAULT '',
    multisig_redeem_script TEXT NOT NULL DEFAULT '',
    multisig_descriptor TEXT NOT NULL DEFAULT '',
    multisig_network TEXT NOT NULL DEFAULT 'regtest',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vaults_status ON vaults(status);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vault_status') THEN
        CREATE TYPE vault_status AS ENUM ('ACTIVE', 'DORMANT');
    END IF;
END$$;

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_vaults_modtime ON vaults;
CREATE TRIGGER update_vaults_modtime
    BEFORE UPDATE ON vaults
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();
