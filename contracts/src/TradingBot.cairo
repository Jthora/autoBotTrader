// TradingBot.cairo (skeleton)
// Cairo 2.x contract skeleton implementing storage layout & basic events per spec.
// Logic (composite calculation, role enforcement, cooldown, etc.) intentionally deferred.

#[starknet::contract]
mod trading_bot {
    use starknet::contract::ContractState;
    use starknet::get_block_timestamp;
    use starknet::info::get_caller_address;

    #[storage]
    struct Storage {
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
        reserved_0: felt252,
        reserved_1: felt252,
    }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct PredictionUpdated {
        astrology: u32,
        gravity: u32,
        ml: u32,
        composite: u32,
        formula_version: u32,
        normalization_version: u32,
    }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct TradeExecuted {
        trade_id: felt252,
        direction: u8,
        amount: u128,
        score: u32,
        timestamp: u64,
        improved_price_bps: u32,
    }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct WeightsUpdated { astrology_w: u32, gravity_w: u32, ml_w: u32 }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct RolesUpdated { pusher_role: felt252, ml_oracle: felt252 }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct ThresholdUpdated { threshold: u32 }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct CooldownUpdated { cooldown_seconds: u32 }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct MLScoreUpdated { ml_score: u32, ml_score_timestamp: u64, ml_model_version: u32 }

    #[event]
    #[derive(Copy, Drop, Serde)]
    struct AdminChanged { new_admin: felt252 }

    #[constructor]
    fn constructor(ref self: ContractState, admin: felt252, pusher_role: felt252) {
        self.admin.write(admin);
        self.pusher_role.write(pusher_role);
        self.ml_oracle.write(0);
        self.execution_threshold.write(50_u32);
        self.cooldown_seconds.write(0_u32);
        self.last_input_timestamp.write(0_u64);
        self.formula_version.write(1_u32);
        self.normalization_version.write(1_u32);
        self.astrology_w.write(50_u32);
        self.gravity_w.write(50_u32);
        self.ml_w.write(0_u32);
        self.astrology_score.write(0_u32);
        self.gravity_score.write(0_u32);
        self.ml_score.write(0_u32);
        self.ml_score_timestamp.write(0_u64);
        self.ml_model_version.write(0_u32);
        self.composite_score.write(0_u32);
    }

    #[view]
    fn get_state(self: @ContractState) -> (execution_threshold: u32, cooldown_seconds: u32, formula_version: u32, normalization_version: u32, composite_score: u32) {
        (
            self.execution_threshold.read(),
            self.cooldown_seconds.read(),
            self.formula_version.read(),
            self.normalization_version.read(),
            self.composite_score.read()
        )
    }

    #[view]
    fn get_weights(self: @ContractState) -> (astrology_w: u32, gravity_w: u32, ml_w: u32) {
        (
            self.astrology_w.read(),
            self.gravity_w.read(),
            self.ml_w.read()
        )
    }

    // Internal: compute composite given scores & weights.
    fn compute_composite(a: u32, g: u32, ml: u32, aw: u32, gw: u32, mw: u32) -> u32 {
        let total = aw + gw + mw;
        if total == 0 { return 0; }
        // (a*aw + g*gw + ml*mw) / total
        let num = a * aw + g * gw + ml * mw;
        return num / total;
    }

    // Internal: ensure caller is pusher role
    fn ensure_pusher(self: @ContractState) {
        if get_caller_address() != self.pusher_role.read() {
            panic_with("NOT_PUSHER");
        }
    }

    // Internal: ensure caller is admin
    fn ensure_admin(self: @ContractState) {
        if get_caller_address() != self.admin.read() {
            panic_with("NOT_ADMIN");
        }
    }

    // Internal: ensure caller ml oracle
    fn ensure_ml_oracle(self: @ContractState) {
        if get_caller_address() != self.ml_oracle.read() {
            panic_with("NOT_ML_ORACLE");
        }
    }

    #[external]
    fn set_prediction_inputs(ref self: ContractState, astrology_score: u32, gravity_score: u32) {
        self.ensure_pusher();
        if astrology_score > 100_u32 { panic_with("ASTRO_RANGE"); }
        if gravity_score > 100_u32 { panic_with("GRAV_RANGE"); }
        let ts: u64 = get_block_timestamp().into();
        let last = self.last_input_timestamp.read();
        let cd = self.cooldown_seconds.read().into();
        if cd > 0_u64 && ts < last + cd { panic_with("COOLDOWN"); }

        self.astrology_score.write(astrology_score);
        self.gravity_score.write(gravity_score);
        let composite = compute_composite(
            astrology_score,
            gravity_score,
            self.ml_score.read(),
            self.astrology_w.read(),
            self.gravity_w.read(),
            self.ml_w.read(),
        );
        self.composite_score.write(composite);
        self.last_input_timestamp.write(ts);
        self.emit(PredictionUpdated { astrology: astrology_score, gravity: gravity_score, ml: self.ml_score.read(), composite, formula_version: self.formula_version.read(), normalization_version: self.normalization_version.read() });
    }

    #[external]
    fn set_ml_score(ref self: ContractState, ml_score: u32, ml_model_version: u32) {
        self.ensure_ml_oracle();
        if ml_score > 100_u32 { panic_with("ML_RANGE"); }
        self.ml_score.write(ml_score);
        self.ml_model_version.write(ml_model_version);
        self.ml_score_timestamp.write(get_block_timestamp().into());
        let composite = compute_composite(
            self.astrology_score.read(),
            self.gravity_score.read(),
            ml_score,
            self.astrology_w.read(),
            self.gravity_w.read(),
            self.ml_w.read(),
        );
        self.composite_score.write(composite);
        self.emit(MLScoreUpdated { ml_score, ml_score_timestamp: self.ml_score_timestamp.read(), ml_model_version });
    }

    #[external]
    fn set_weights(ref self: ContractState, astrology_w: u32, gravity_w: u32, ml_w: u32) {
        self.ensure_admin();
        if astrology_w + gravity_w + ml_w == 0_u32 { panic_with("ZERO_TOTAL_W"); }
        self.astrology_w.write(astrology_w);
        self.gravity_w.write(gravity_w);
        self.ml_w.write(ml_w);
        let composite = compute_composite(
            self.astrology_score.read(),
            self.gravity_score.read(),
            self.ml_score.read(),
            astrology_w,
            gravity_w,
            ml_w,
        );
        self.composite_score.write(composite);
        self.emit(WeightsUpdated { astrology_w, gravity_w, ml_w });
    }

    #[external]
    fn set_threshold(ref self: ContractState, threshold: u32) {
        self.ensure_admin();
        if threshold > 100_u32 { panic_with("THRESH_RANGE"); }
        self.execution_threshold.write(threshold);
        self.emit(ThresholdUpdated { threshold });
    }

    #[external]
    fn set_cooldown(ref self: ContractState, cooldown_seconds: u32) {
        self.ensure_admin();
        self.cooldown_seconds.write(cooldown_seconds);
        self.emit(CooldownUpdated { cooldown_seconds });
    }

    #[external]
    fn set_roles(ref self: ContractState, pusher_role: felt252, ml_oracle: felt252) {
        self.ensure_admin();
        self.pusher_role.write(pusher_role);
        self.ml_oracle.write(ml_oracle);
        self.emit(RolesUpdated { pusher_role, ml_oracle });
    }

    #[external]
    fn transfer_admin(ref self: ContractState, new_admin: felt252) {
        self.ensure_admin();
        self.admin.write(new_admin);
        self.emit(AdminChanged { new_admin });
    }
}
