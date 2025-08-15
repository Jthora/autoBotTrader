// compute.cairo - pure helper function extracted for isolated testing.

pub fn composite(a: u32, g: u32, ml: u32, aw: u32, gw: u32, mw: u32) -> u32 {
    let total = aw + gw + mw;
    if total == 0 { return 0; }
    (a * aw + g * gw + ml * mw) / total
}
