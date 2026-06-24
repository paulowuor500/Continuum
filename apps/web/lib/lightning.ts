export function compactInvoice(invoice: string): string {
  if (invoice.length <= 34) {
    return invoice;
  }
  return `${invoice.slice(0, 18)}...${invoice.slice(-12)}`;
}

export function demoInvoice(vaultId: string): string {
  return `lnbc1u1continuum_demo_${vaultId.replaceAll("-", "").slice(0, 18)}_proof_of_life`;
}
