import { compactInvoice } from "../lib/lightning";

export default function LightningQR({ invoice }: { invoice: string }) {
  return (
    <div className="panel" style={{ display: "grid", gap: 14 }}>
      <div className="panel-header" style={{ marginBottom: 0 }}>
        <div>
          <p className="eyebrow">Lightning invoice</p>
          <h3>Proof payment</h3>
        </div>
      </div>
      <div
        aria-label="Stylized invoice QR placeholder"
        style={{
          aspectRatio: "1",
          background:
            "repeating-linear-gradient(90deg, #f2f5f4 0 8px, #101112 8px 16px), repeating-linear-gradient(0deg, rgba(16,17,18,.55) 0 10px, transparent 10px 20px)",
          border: "10px solid #f2f5f4",
          maxWidth: 180,
          width: "100%"
        }}
      />
      <p className="mono truncate">{compactInvoice(invoice)}</p>
    </div>
  );
}
