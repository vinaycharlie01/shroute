# OmniRouter Frontend

This directory is a scaffold for the OmniRouter frontend. The build tooling
(`npm` scripts + the `Frontend` Mage namespace at the repo root, configured by
`node.yaml`) is wired up, but the actual application code has not been moved
here yet.

## Status

The existing Next.js application lives at `../OmniRoute/` and is **not**
touched by this milestone. It remains the reference implementation while the
Go backend foundation (`../backend/`) is established. Migrating it into this
directory is a separate, future piece of work.

## Planned structure

Once migration starts, this directory is expected to hold a self-contained
Next.js app, independent of the backend:

```
frontend/
  src/            # application source
  public/         # static assets
  package.json    # npm scripts: dev, build, lint, test
```

## Build tooling

No Makefiles or shell scripts: every frontend dev/CI task runs through Mage
(`nava/mage/nodejs`), driven by `node.yaml` at the repo root.

```bash
mage frontend:setup   # npm install (node.yaml -> setup)
mage frontend:dev     # npm run dev (node.yaml -> dev)
mage frontend:build   # npm run build (node.yaml -> build)
mage frontend:lint    # npm run lint (node.yaml -> lint)
mage frontend:test    # npm run test (node.yaml -> test)
```

The `npm` scripts in `package.json` are currently stubs; they will be
replaced with real Next.js scripts when the application is migrated.

## Migration plan (future milestone)

1. Move (not copy) the relevant application code from `../OmniRoute/` into
   `frontend/`, preserving git history where practical.
2. Replace the stub `package.json` scripts with the real Next.js
   dev/build/lint/test commands.
3. Point the frontend's API client at the Go backend's endpoints.
4. Wire `mage frontend:build` into CI alongside the backend build.
5. Decommission `../OmniRoute/` once the migration is verified.
