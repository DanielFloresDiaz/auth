# Changelog

## [0.1.0] - 2026-10-12

This release marks the first versioned distribution of the maintained auth fork: upstream GoTrue/Supabase Auth capabilities plus Solomon-specific authentication, administration, and operations improvements.

**Passkeys and WebAuthn:**

* Add end-user passkey registration and authentication flows, including management APIs for listing and removing credentials.
* Expose admin endpoints to inspect and manage user passkeys where operators need support and audit visibility.

**MFA and recovery codes:**

* Ship recovery-code generation, verification, regeneration, and deletion alongside existing TOTP factors.
* Extend configuration and metering for recovery-code usage so operators can enable the feature safely in production.

**SCIM provisioning:**

* Introduce SCIM v2 foundations: service provider configuration, resource types, schemas, and RFC 7644 error responses.
* Add persistence and routing for SCIM users and tokens behind an experimental feature flag.

**OAuth and external identity:**

* Support custom OAuth providers with admin configuration and runtime provider integration.
* Harden the embedded OAuth authorization server with improved client authentication, metering, and replay-safety coverage.
* Expand provider coverage and OIDC caching behavior for external sign-in reliability.

**SAML and enterprise SSO:**

* Improve SAML assertion consumer handling and document SAML key rotation for operators rotating IdP or SP keys.

**Configuration, secrets, and crypto:**

* Refactor configuration loading with structured env parsing and safer conf loading paths for large deployments.
* Support AWS KMS–backed RSA JWK material for JWT signing where keys must not live on disk.
* Extend password, recovery-code, and related crypto primitives used by new MFA features.

**API quality, limits, and observability:**

* Add request rate limiting middleware for abusive traffic patterns.
* Improve pagination, token verification (including one-time token expiry parity), and shared JSON/HTTP helpers used across handlers.
* Refresh audit, identity, and admin behaviors around passwords, email identities, and provider constants.

**Operations:**

* Align container and CI release pipelines with tagged binaries and images; apply database migrations bundled with release artifacts as before.
* Pin and verify Go toolchain versions across Dockerfiles and local builds for reproducible releases.
