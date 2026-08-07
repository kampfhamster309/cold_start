import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Standalone output for the Docker build (tech-stack §1) — bundles only
  // what's needed to run into .next/standalone, no full node_modules copy.
  output: "standalone",
};

export default nextConfig;
