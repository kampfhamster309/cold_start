import type { Metadata } from "next";
import "./globals.css";

// Deliberately no next/font/google here: this is a self-hosted product
// (tech-stack §8), and a default that fetches fonts from Google at build
// time works against that from the very first page. System font stack
// instead — no external dependency, no build-time network requirement.

export const metadata: Metadata = {
  title: "cold_start",
  description: "Self-hosted onboarding platform",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
