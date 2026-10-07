# Retired Authorization Service

This repository previously ran Traefik ForwardAuth for every protected API request, verified Gauas access JWTs, consulted Redis for revoked session IDs, and injected `X-Gauas-*` identity headers. That service has been retired from the target architecture.

Identity Service now owns login, sessions, token signing and JWKS. HunterJob verifies Bearer JWTs locally with cached JWKS and keeps application authorization in its own domain. Traefik routes requests without ForwardAuth. Logout revokes refresh sessions; issued access tokens expire after a short lifetime.

The old source, container build and image workflow were removed from this repo. The Git history retains them if an audit or rollback is required. Do not deploy this repository's old image during migration.
