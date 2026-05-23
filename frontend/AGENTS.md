# Frontend Guidelines

## Project Structure & Module Organization

This directory is the React/Vite frontend for `jcourse`. Application code lives in `src/`. Pages are in `src/pages`, API clients in `src/api`, hooks in `src/hooks`, shared UI primitives in `src/components/ui`, feature components in `src/components/*`, routing in `src/router.tsx`, and utility helpers in `src/lib`. Static assets live in `public/`, and MSW workers are generated there when mocks are used.

## Build, Test, and Development Commands

- `pnpm dev` starts the Vite development server.
- `pnpm build` runs the production build and type-checks the app.
- `pnpm lint` runs ESLint across the project.
- `pnpm typecheck` runs TypeScript without bundling.
- `pnpm test` runs Vitest with `--passWithNoTests`.
- `pnpm preview` serves the built app locally.

## Coding Style & Naming Conventions

Use TypeScript, React function components, and the existing path aliases such as `@/components/ui/button`. Keep component names in PascalCase, hooks as `useThing`, and page files in kebab-case such as `course-detail-page.tsx`. Prefer existing UI primitives and feature folders before introducing new abstractions. Keep state, API calls, and formatting logic close to the feature that uses them.

## Testing Guidelines

Frontend changes should pass `pnpm lint` and `pnpm typecheck`. Add focused Vitest coverage for nontrivial hooks, client behavior, and UI logic when a change carries real branching or data transformation. Keep tests near the code they exercise when practical.

## Commit & Pull Request Guidelines

Keep commits short and imperative, with an area or type when useful. Pull requests should summarize the user-visible change, list commands run, and include screenshots for UI updates.

## Security & Configuration Tips

Do not commit real credentials, API tokens, or environment-specific overrides. Preserve the existing API contract when changing frontend requests or route handling, and verify that new UI paths still work with the current auth flow.
