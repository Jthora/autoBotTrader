# Expanded Test Plan

This document enumerates detailed test specifications with subtests for each component. It augments `docs/initialPlan/09_testing_strategy.md`.

## 1. Scope & Goals

Provide exhaustive functional, edge, negative, concurrency, fuzz, and governance tests ensuring:

- Deterministic correctness of core calculations.
- Robust handling of malformed inputs & partial failures.
- Concurrency safety (race detector clean).
- Early detection of placeholder / stub leakage into CI.

## 2. Component Matrix

| Component           | Languages | Critical Risk Areas                               |
| ------------------- | --------- | ------------------------------------------------- |
| GTAB Loader         | Go        | Binary parsing, boundary interpolation            |
| FileGravimetric     | Go        | Concurrency, hysteresis correctness, stale window |
| Normalization       | Go        | Clamping, non-finite handling, precision          |
| Chain Client        | Go        | Env gating, RPC failure modes, context timeouts   |
| HTTP Handlers       | Go        | Error propagation, dry-run vs real path           |
| Progress Automation | Python    | Graph integrity, duplicate IDs, logging           |
| Gas Snapshot        | Python    | Placeholder governance                            |
| Kernel Verification | Python    | Hash placeholder lifecycle                        |
| Concurrency/Fuzz    | Go        | Data races, invariant preservation                |
| Cairo Contract      | Cairo     | Access control, cooldown, composite math          |

## 3. Test Taxonomy

- Functional: Happy path logic.
- Edge: Bounds, limits, boundary conditions.
- Negative: Expected failures / reverts.
- Concurrency: Parallel execution & race detection.
- Fuzz / Property: Randomized invariant testing.
- Governance: Ensuring placeholders / stubs cannot silently pass.

## 4. Detailed Specifications

### 4.1 GTAB Loader (`api/internal/ephem/gtab.go`)

Tests (file: `gtab_test.go`):

- `TestOpenBadMagic`
  - wrong_magic_raises_error
  - unsupported_version_error
  - zero_dt_rejected
  - zero_n_rejected
- `TestLookupTideBPSBoundaries`
  - exact_start_index
  - exact_end_index
  - before_start_out_of_range
  - after_end_out_of_range
- `TestInterpolationRounding`
  - midpoint_rounds_closer
  - biased_up_round
  - biased_down_round
  - near_65535_clamped
- `TestMissingTideField`
  - fields_mask_excludes_tide
  - mixed_fields_size_validation
  - access_after_close
  - malformed_truncated_file

Helper: `makeTempGTAB(samples int, dt time.Duration, fieldsMask uint32, values []uint16)`.

### 4.2 FileGravimetric Provider (`api/internal/providers/gravity_file.go`)

Tests (extend / new):

- `TestHysteresisBehavior`
  - within_band_sticky
  - edge_equal_hysteresis
  - reset_after_unset
  - direction_flip_small_delta
- `TestOutOfRangeClamping`
  - before_start_uses_start_value
  - after_end_uses_end_value
  - both_missing_defaults_zero
  - stale_flag_logic
- `TestConcurrentSafety` (expand existing race test)
  - parallel_fetch_only
  - mixed_fetch_and_fetchAt
  - cancel_context_mid_fetch
  - high_volume_iterations
- `TestDatasetIDResolution`
  - explicit_dataset_id_param
  - meta_file_present_sets_id
  - meta_file_missing_leaves_empty
  - malformed_meta_file_ignored

### 4.3 Normalization (`api/internal/normalize/normalize.go`)

- `TestAstrologyScoreBounds`: below_min_clamped, above_max_clamped, exact_min_zero, exact_max_hundred
- `TestGravimetricScoreBounds`: same subtests
- `TestNonFiniteInputs`: nan_returns_zero, inf_clamped_hundred, neg_inf_clamped_zero, huge_positive_clamped
- `TestPrecisionSteps`: mid_value_rounding, near_boundary_down, near_boundary_up, idempotent_same_input

### 4.4 Chain Client (`api/internal/chain/client.go`)

- `TestDisabledConfigMockPath`: missing_account_address_disables, only_rpc_and_contract_mock, full_env_required_enables, push_returns_mock_hash
- `TestPushPredictionValidation`: astro_above_100_error, grav_above_100_error, both_valid_ok, context_cancelled_before_request
- `TestRPCFailureClassification` (httptest server): non_200_status_error, timeout_error, malformed_json_body_ignored, network_refused_error
- `TestGetCompositeStub`: disabled_returns_zero, enabled_rpc_success_returns_zero, rpc_error_propagates, multiple_calls_no_state_leak

### 4.5 HTTP Handlers (`api/internal/http/handlers.go`)

- `TestPredictWeightsAndComposite`: equal_inputs_equal_output, different_inputs_average, provider_latency_within_timeout, provider_failure_returns_error
- `TestPushEndpoint`: dry_run_when_chain_nil, real_push_when_PUSH_REAL, push_score_validation_error, provider_fetch_error_handled
- `TestHealthMetadata`: includes_mode_dataset, stale_flag_true, timestamp_rfc3339nano, responds_fast
- `TestErrorResponses`: gravimetrics_fetch_failure_503, astrology_fetch_failure_503, predict_astrology_failure_short_circuit, push_chain_error_no_500
- `TestParallelPredictAndPush` (stress): 20_parallel_predict, predict_push_mix, context_cancel_mid_mix, no_data_races

### 4.6 Progress Automation (`scripts/progress/update_progress.py`)

- `TestTaskGraphValidation`: missing_dependency_detected, cyclic_dependency_detected, invalid_status_flagged, duplicate_ids_flagged
- `TestReadyComputation`: all_deps_complete_ready, missing_dep_not_ready, blocked_excluded, priority_sort_order
- `TestMetricsRecompute`: counts_after_status_change, timestamp_changes, empty_graph_safe, metrics_no_mutation_original
- `TestStatusLogAppend`: diff_line_present_on_change, placeholder_on_no_change, malformed_previous_metrics_safe, ready_queue_included

### 4.7 Gas Snapshot (`scripts/progress/gas_snapshot.py`)

- `TestSnapshotGeneration`: writes_file_with_functions, overwrite_existing, iso_timestamp_present, non_zero_enforcement_flag
- `TestPlaceholderEnforcement`: zeros_fail_without_flag, zeros_allowed_with_env, mixed_zero_nonzero_fails, all_nonzero_passes

### 4.8 Kernel Verification (`scripts/ephem/validation/verify_kernel.py`)

- `TestPlaceholderUpdateFlow`: placeholder_detected_without_update, update_rewrites_hash, second_update_idempotent, corrupted_file_rejects

### 4.9 Concurrency & Fuzz

- `TestParallelPredictAndPush` (already listed under handlers)
- Fuzz target `FuzzLookupTideBPS` (Go fuzz): invariants (never >65535, ok range consistency, monotonic index mapping)
- Property (optional) normalization monotonicity

### 4.10 Cairo Contract (snforge tests)

- `test_set_prediction_inputs_validation`: astrology_score_above_100_reverts, gravity_score_above_100_reverts, valid_scores_event_emitted, cooldown_blocks_second_call
- `test_set_weights`: zero_sum_reverts, update_changes_storage, composite_recomputed, event_emitted
- `test_ml_score`: ml_score_above_100_reverts, update_sets_timestamp, composite_recomputed, event_emitted
- `test_execute_trade`: below_threshold_reverts, at_threshold_succeeds, event_fields_correct, direction_invalid_reverts
- `test_role_access`: non_admin_set_weights_reverts, non_pusher_set_prediction_inputs_reverts, non_oracle_set_ml_score_reverts, admin_change_updates_admin

## 5. Fixtures & Utilities

- GTAB mini-generator helper (Go): writes in-memory file; store under `api/testutil`.
- RPC mock server: configurable responses & latency.
- Temp task graph builder for Python tests.
- Small static GTAB + meta in `tests/testdata/gtab/` (≤5KB) or generated in setup.

## 6. Implementation Order

1. Core Safety: GTAB, FileGravimetric, Chain validation, Predict/Push base.
2. Governance: Progress duplicate ID, Gas placeholder enforcement, Kernel update flow.
3. Contract Stateful suite.
4. Edge & Fuzz (normalization non-finite, fuzz lookup).
5. Performance & Benchmark gating tests.

## 7. CI Hooks

- Add `go test -race ./...` (already) + separate `go test -run Fuzz -fuzztime=5s` job (opt-in).
- Add Python unit tests via `pytest` (introduce if not present) for scripts.
- Add placeholder scan step: `scripts/ci/check_placeholders.py` (future).

## 8. Placeholder Fail Policy

Critical files (gas_snapshot.json, hashes manifest) must not contain tokens: PLACEHOLDER, \_PLACEHOLDER unless environment `ALLOW_PLACEHOLDERS=1`.

## 9. Coverage Targets

| Layer                                      | Target |
| ------------------------------------------ | ------ |
| Go core logic (ephem/providers/chain/http) | ≥80%   |
| Cairo contract critical functions          | ≥85%   |
| Python governance scripts                  | ≥70%   |

## 10. Future Extensions

- Contract fuzz harness when available.
- Event-driven integration test (local devnet) for end-to-end composite change.
- Performance regression snapshot diffs (bench harness).

---

This plan should be treated as living; update timestamps & rationale when modifying scope.

## Aug 2025 Effort Sizing Addendum

Relative effort (S/M/L) for remaining major items (see roadmap & task graph):

- Gas extraction parser integration: S (0.5–1 day)
- Contract unit tests (stateful core suite): M (1.5–2.5 days)
- Accuracy harness (compute_metrics.py + nightly wiring): M/L (2–3.5 days)
- Real Starknet chain client (tx send + composite read): M (1.5–2.5 days)
- Frontend scaffold & score display: S (0.5–1 day)

Priority / ROI Order:

1. Gas parser (unblocks governance)
2. Contract tests (baseline correctness + future gas stability)
3. Chain client (enables end-to-end value)
4. Frontend scaffold (user visibility)
5. Accuracy harness (scientific validation; can run parallel once others in motion)

These sizes guide sprint slicing and task decomposition; adjust if scope shifts (e.g., adding signature verification to chain client would upgrade it to L).
