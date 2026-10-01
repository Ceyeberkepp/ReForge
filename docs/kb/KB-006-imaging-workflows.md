# KB-006 — Gold, Clone, ISO, and Capture workflows

## Install sources

Authenticated PXE users can be granted access to:

- ISO
- Gold Image
- Clone Image

## Capture destinations

Authorized users can request:

- Gold Image capture
- Clone Image capture

## Gold Image

Use for standardized reusable operating-system images, normally generalized before deployment.

## Clone Image

Use for machine/family-style captured images where the deployment workflow intentionally preserves more source-machine state.

## ISO

Use original installation/recovery media as the deployment source.

## Important current limitation

ReForge currently provides the control-plane authorization and task orchestration, but destructive raw-disk restore/capture requires the dedicated imaging node.

Until that imaging node performs the operation, the task should remain waiting rather than showing false completion.

This distinction is required for support accuracy.
