import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Continuum",
  description: "Privacy-preserving Bitcoin inheritance with Lightning proof-of-life."
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
