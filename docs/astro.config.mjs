import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import mdx from "@astrojs/mdx";
import expressiveCode from "astro-expressive-code";

export default defineConfig({
  site: "https://mcptocli.com",
  output: "static",
  integrations: [
    expressiveCode({
      themes: ["dark-plus"],
      frames: {
        showCopyToClipboardButton: true,
        extractFileNameFromCode: false,
      },
      defaultProps: {
        frame: "code",
      },
    }),
    mdx(),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
