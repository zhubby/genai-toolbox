import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    // 将前端 /tb-api/* 代理到后端 http://127.0.0.1:5000/*，避免 CORS
    return [
      {
        source: "/tb-api/:path*",
        destination: "http://127.0.0.1:5000/:path*",
      },
    ];
  },
};

export default nextConfig;
