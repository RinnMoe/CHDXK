import { defineConfig } from "oxlint"

export default defineConfig({
  env: {
    browser: true,
    es2020: true,
  },
  ignorePatterns: ["dist", "public/mockServiceWorker.js"],
  jsPlugins: ["oxlint-tailwindcss"],
  plugins: ["typescript", "react", "oxc", "unicorn", "react-perf"],
  settings: {
    react: {
      version: "19.2.6",
    },
    "tailwindcss" :{
      "entryPoint": "src/index.css"
    }
  },
  options: {
    typeAware: true,
  },
  rules: {
    "tailwindcss/no-unknown-classes": "error",
    "tailwindcss/no-duplicate-classes": "error",
    "tailwindcss/no-conflicting-classes": "error",
    "tailwindcss/no-deprecated-classes": "error",
    "tailwindcss/no-unnecessary-whitespace": "error",
    "tailwindcss/enforce-sort-order": "warn",
    "tailwindcss/enforce-shorthand": "warn",
    "tailwindcss/enforce-consistent-variable-syntax": "warn",
    "tailwindcss/enforce-canonical": "warn",
    "tailwindcss/enforce-consistent-important-position": "warn",
    "tailwindcss/no-hardcoded-colors": "warn",
    "tailwindcss/prefer-theme-tokens": "warn",
    "max-lines": ["error", { max: 400 }],
    "react/exhaustive-deps": "error",
    "react/only-export-components": "error",
    "react/rules-of-hooks": "error",
    "typescript/no-deprecated": "error",
  },
  overrides: [
    {
      files: ["src/components/ui/**/*.{ts,tsx}", "src/contexts/**/*.{ts,tsx}"],
      rules: {
        "react/only-export-components": "off",
      },
    },
  ],
})
