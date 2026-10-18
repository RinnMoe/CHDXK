# ResHub refresh provider handoff (2026-10-19)

Status: local implementation committed and verified; production untouched.

Provider worktree: `C:/Users/Rinn/.codex/worktrees/reshub-native-bridge/res-hub`.
Branch: `codex/native-chdxk-bridge`, clean baseline `9ce58f2` (native bridge).
Master baseline is `4d2455b`; native baseline is one commit ahead. Final refresh
commit is `6d31a08` (feat: 支持受父会话约束的 OIDC 刷新令牌轮换). Future release
MUST include BOTH native bridge `9ce58f2` and refresh `6d31a08`, in that order.

## Confirmed contract

- `POST /oidc/token`, form encoding; `client_secret_basic` and existing TLS/mTLS.
- Send `grant_type=refresh_token`, `refresh_token`; omit scope to keep original.
- Response: access_token, token_type=Bearer, expires_in, refresh_token, scope.
- No id_token on refresh: retain original id_token for logout.
- Strict single-use rotation. Replaying a consumed token returns 400 invalid_grant
  AND revokes that issuance chain's access/refresh tokens, including the winner
  of a concurrent request. CHDXK must serialize per session across all instances.
- No retry/grace window: successful provider response lost in transport means
  fresh login. A local shared result cache cannot recover a response never received.
- No scope expansion/new scopes. Requested narrowed scopes persist on rotation.
- Only digests are stored. Consume + access + successor insertion is transactional.

## TTL and compatibility

Access lifetime defaults to 3600 seconds, capped by the original refresh-chain
absolute expiry. Refresh expiry equals the original ResHub login expiry, capped
by operator/parent session expiry; rotating never extends it. There is no extra
OIDC refresh TTL. Host sessions normally expire after seven days.

Existing clients require an admin update: append `refresh_token` to allowed_grants,
retaining authorization_code, existing service grants, scopes and callbacks.
New clients likewise opt in explicitly. No schema/client-row backfill grants
refresh automatically. A fresh browser/native authorization-code exchange then
includes refresh_token. Already existing sessions without refresh_token continue
with old behavior until a fresh login. Client credentials do not issue refresh.
Native bridge still requires OIDC_NATIVE_CLIENT_IDS allowlist.

Revocation: host logout, password-change/session revocation, deleted users,
invalid parent sessions, client disabled/secret rotated, expired tokens all stop
refresh. `/oidc/revoke` accepts a refresh token and revokes its issuance chain.
There is no standalone ban field in current ResHub: account bans must revoke
subject/session; violation_count by itself is not a ban.

## Verification so far

Final combined OIDC/native/refresh regression: **47 passed** (2026-10-19).
This includes 18 new integration/storage cases. AST parsing of 5 affected Python
files and git diff --check passed. Provider worktree is clean after commit.

Tests run using an ephemeral test-only SECRET_KEY and a process-local webadmin
package alias pointing to this worktree (repository folder is named res-hub).
No test secret was saved in the repository. No deployment, production access,
or real browser/device login was performed.

## Local cross-service fixture

Real main.app with isolated temporary SQLite, synthetic client and user, no-op
lifespan, loopback-only HTTP is available for CHDXK Go adapter validation.
Credentials are shared only through a temporary JSON outside Git.

CHDXK confirmed this real-adapter command passed:
`go test -tags test ./internal/infrastructure/reshub -run '^TestRefreshAgainstLocalResHub$' -v -count=1`
with JCOURSE_TEST_OIDC_FIXTURE set to the temporary fixture. It exercised real
Discovery -> Refresh -> UserInfo subject binding -> second rotation -> first
refresh replay invalid_grant -> newest access token rejected by UserInfo.
CHDXK also confirmed the real middleware/Redis test passed:
`go test -race -tags test ./internal/interface/web/middleware -run '^TestOIDCRefreshAgainstLocalProviderAndRedis$' -v -count=1`
using the second independent synthetic refresh chain and isolated localhost
Redis. It exercised real Python app/SQLite -> Go refresh coordinator -> Redis
session -> identity resolution -> exactly one business write -> logout -> old
Cookie business request rejected with 401.

Fixture cleanup completed: verified Python fixture PID stopped; temporary
synthetic credentials, SQLite files and runner directory sent to Windows Recycle
Bin. Both failed fixture attempts were likewise recycled. No fixture credentials
are recorded in this handoff or Git. The provider repository remains clean at
6d31a08; no push, deployment, or production operation was performed.
