// In-memory simulation harness for trading_bot storage & logic (no Starknet syscalls).
// Uses free functions because inherent impls are avoided for simplicity.

#[derive(Copy, Drop)]
struct SimState {
    admin: felt252,
    pusher_role: felt252,
    ml_oracle: felt252,
    execution_threshold: u32,
    cooldown_seconds: u32,
    last_input_timestamp: u64,
    formula_version: u32,
    normalization_version: u32,
    astrology_w: u32,
    gravity_w: u32,
    ml_w: u32,
    astrology_score: u32,
    gravity_score: u32,
    ml_score: u32,
    ml_score_timestamp: u64,
    ml_model_version: u32,
    composite_score: u32,
    trade_count: u32,
}

fn sim_new(admin: felt252, pusher: felt252) -> SimState {
    SimState {
        admin,
        pusher_role: pusher,
        ml_oracle: admin,
        execution_threshold: 50_u32,
        cooldown_seconds: 0_u32,
        last_input_timestamp: 0_u64,
        formula_version: 1_u32,
        normalization_version: 1_u32,
        astrology_w: 50_u32,
        gravity_w: 50_u32,
        ml_w: 0_u32,
        astrology_score: 0_u32,
        gravity_score: 0_u32,
        ml_score: 0_u32,
        ml_score_timestamp: 0_u64,
        ml_model_version: 0_u32,
        composite_score: 0_u32,
        trade_count: 0_u32,
    }
}

fn sim_compute_composite(a: u32, g: u32, ml: u32, aw: u32, gw: u32, mw: u32) -> u32 {
    let total = aw + gw + mw;
    if total == 0 { return 0; }
    (a * aw + g * gw + ml * mw) / total
}

fn sim_set_prediction_inputs(ref state: SimState, caller: felt252, ts: u64, astrology_score: u32, gravity_score: u32) {
    assert(caller == state.pusher_role, 'NOT_PUSHER');
    assert(astrology_score <= 100_u32, 'INVALID_SCORE');
    assert(gravity_score <= 100_u32, 'INVALID_SCORE');
    let cd: u64 = state.cooldown_seconds.into();
    if cd > 0_u64 && ts < state.last_input_timestamp + cd { assert(false, 'COOLDOWN'); }
    state.astrology_score = astrology_score;
    state.gravity_score = gravity_score;
    state.composite_score = sim_compute_composite(astrology_score, gravity_score, state.ml_score, state.astrology_w, state.gravity_w, state.ml_w);
    state.last_input_timestamp = ts;
}

fn sim_set_ml_score(ref state: SimState, caller: felt252, ts: u64, ml_score: u32, ml_model_version: u32) {
    assert(caller == state.ml_oracle, 'NOT_ORACLE');
    assert(ml_score <= 100_u32, 'INVALID_SCORE');
    state.ml_score = ml_score;
    state.ml_model_version = ml_model_version;
    state.ml_score_timestamp = ts;
    state.composite_score = sim_compute_composite(state.astrology_score, state.gravity_score, ml_score, state.astrology_w, state.gravity_w, state.ml_w);
}

fn sim_set_weights(ref state: SimState, caller: felt252, astrology_w: u32, gravity_w: u32, ml_w: u32) {
    assert(caller == state.admin, 'NOT_ADMIN');
    assert(astrology_w + gravity_w + ml_w > 0_u32, 'ZERO_WEIGHTS');
    state.astrology_w = astrology_w;
    state.gravity_w = gravity_w;
    state.ml_w = ml_w;
    state.composite_score = sim_compute_composite(state.astrology_score, state.gravity_score, state.ml_score, astrology_w, gravity_w, ml_w);
}

fn sim_update_threshold(ref state: SimState, caller: felt252, threshold: u32) {
    assert(caller == state.admin, 'NOT_ADMIN');
    state.execution_threshold = threshold;
}

fn sim_update_cooldown(ref state: SimState, caller: felt252, cooldown_seconds: u32) {
    assert(caller == state.admin, 'NOT_ADMIN');
    state.cooldown_seconds = cooldown_seconds;
}

fn sim_set_roles(ref state: SimState, caller: felt252, pusher_role: felt252, ml_oracle: felt252) {
    assert(caller == state.admin, 'NOT_ADMIN');
    state.pusher_role = pusher_role;
    state.ml_oracle = ml_oracle;
}

fn sim_change_admin(ref state: SimState, caller: felt252, new_admin: felt252) {
    assert(caller == state.admin, 'NOT_ADMIN');
    state.admin = new_admin;
}

fn sim_execute_trade(ref state: SimState, trade_id: felt252, _caller: felt252, _amount: u128, direction: u8, _improved_price_bps: u32) {
    assert(direction == 0_u8 || direction == 1_u8, 'BAD_DIRECTION');
    assert(state.composite_score >= state.execution_threshold, 'LOW_SCORE');
    state.trade_count += 1_u32;
    let _unused = trade_id; // silence
}

// --- Tests ---

const ADMIN: felt252 = 0xAAA;
const PUSHER: felt252 = 0xBBB;
const OTHER: felt252 = 0xCCC;
const ORACLE: felt252 = 0xDDD;

#[test]
fn test_constructor_defaults() {
    let state = sim_new(ADMIN, PUSHER);
    assert(state.execution_threshold == 50_u32, 'bad_threshold');
    assert(state.astrology_w + state.gravity_w + state.ml_w == 100_u32, 'bad_weights');
}

#[test]
fn test_set_prediction_updates_composite() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_set_prediction_inputs(ref state, PUSHER, 1000_u64, 60_u32, 40_u32);
    // weights 50/50 so composite should be avg
    assert(state.composite_score == 50_u32, 'bad_composite');
}

#[test]
#[should_panic(expected: 'COOLDOWN')]
fn test_cooldown_enforced() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_update_cooldown(ref state, ADMIN, 100_u32);
    sim_set_prediction_inputs(ref state, PUSHER, 1000_u64, 10_u32, 10_u32);
    // second call inside cooldown should panic
    sim_set_prediction_inputs(ref state, PUSHER, 1005_u64, 20_u32, 20_u32);
}

#[test]
fn test_set_ml_score_recomputes() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_set_weights(ref state, ADMIN, 40_u32, 40_u32, 20_u32);
    sim_set_prediction_inputs(ref state, PUSHER, 1000_u64, 50_u32, 50_u32); // composite 50 initially (ml 0)
    sim_set_ml_score(ref state, ADMIN, 1010_u64, 100_u32, 1_u32); // admin is oracle default
    // new composite = (50*40 + 50*40 + 100*20)/100 = (2000 + 2000 + 2000)/100 = 6000/100 = 60
    assert(state.composite_score == 60_u32, 'ml_not_applied');
}

#[test]
#[should_panic(expected: 'NOT_ADMIN')]
fn test_roles_enforced_weights() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_set_weights(ref state, OTHER, 10_u32, 10_u32, 80_u32); // should panic
}

#[test]
fn test_execute_trade_threshold() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_set_prediction_inputs(ref state, PUSHER, 1000_u64, 90_u32, 90_u32); // composite 90
    sim_update_threshold(ref state, ADMIN, 80_u32);
    sim_execute_trade(ref state, 1_felt252, OTHER, 100_u128, 1_u8, 10_u32);
    assert(state.trade_count == 1_u32, 'trade_not_recorded');
}

#[test]
#[should_panic(expected: 'LOW_SCORE')]
fn test_execute_trade_low_score_fails() {
    let mut state = sim_new(ADMIN, PUSHER);
    sim_update_threshold(ref state, ADMIN, 60_u32);
    sim_set_prediction_inputs(ref state, PUSHER, 1000_u64, 40_u32, 40_u32); // composite 40 < 60
    sim_execute_trade(ref state, 2_felt252, OTHER, 10_u128, 0_u8, 5_u32);
}
