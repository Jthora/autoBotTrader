// PURE TESTS ONLY (stateful tests temporarily removed pending updated harness APIs)

fn composite(a: u32, g: u32, ml: u32, aw: u32, gw: u32, mw: u32) -> u32 {
    let total = aw + gw + mw;
    if total == 0 { return 0; }
    (a * aw + g * gw + ml * mw) / total
}

#[test]
fn composite_zero_weights_returns_zero() {
    let result = composite(10_u32, 20_u32, 30_u32, 0_u32, 0_u32, 0_u32);
    assert(result == 0_u32, 'non_zero_for_zero_weights');
}

#[test]
fn composite_simple_two_inputs() {
    // a=50 g=70 weights 50/50 no ml
    let result = composite(50_u32, 70_u32, 0_u32, 50_u32, 50_u32, 0_u32);
    assert(result == 60_u32, 'bad_two_input_average');
}

#[test]
fn composite_with_ml() {
    // a=50 g=70 ml=80 weights 40/40/20 expected 64
    let result = composite(50_u32, 70_u32, 80_u32, 40_u32, 40_u32, 20_u32);
    assert(result == 64_u32, 'bad_three_input_weighted');
}

#[test]
fn composite_all_equal_any_weights_returns_same() {
    let result = composite(55_u32, 55_u32, 55_u32, 30_u32, 30_u32, 40_u32);
    assert(result == 55_u32, 'bad_equal_invariance');
}

// TODO: Reintroduce stateful tests (access control, cooldown, trade execution) once test harness updated for Starknet 2.12.
