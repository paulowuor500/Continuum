DROP TABLE IF EXISTS vaults;
DROP TYPE IF EXISTS vault_status;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE vault_status AS ENUM ('ACTIVE', 'DORMANT');

CREATE TABLE vaults (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alias VARCHAR(255) NOT NULL,
    beneficiary_pubkey TEXT NOT NULL,
    encrypted_payload TEXT NOT NULL,          
    check_in_interval_seconds INT NOT NULL,  
    last_check_in_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status vault_status NOT NULL DEFAULT 'ACTIVE',
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vaults_status ON vaults(status);

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_vaults_modtime
    BEFORE UPDATE ON vaults
    FOR EACH ROW
    EXECUTE PROCEDURE update_modified_column();
