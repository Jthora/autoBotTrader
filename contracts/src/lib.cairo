// lib.cairo - simplified to compile under Cairo 2.12; stricter role & error handling can be reintroduced later.

// compute helper kept local for simpler testing

#[starknet::contract]
mod trading_bot {
    use starknet::get_block_timestamp;
    use starknet::contract_address::ContractAddress;
    use starknet::info::get_caller_address;
    // (no external compute module)

    #[storage]
    struct Storage {
        admin: ContractAddress,
        pusher_role: ContractAddress,
        ml_oracle: ContractAddress,
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
    }

    // Event payload structs (no starknet::Event derive here)
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct PredictionUpdated { astrology: u32, gravity: u32, ml: u32, composite: u32, formula_version: u32, normalization_version: u32 }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct WeightsUpdated { astrology_w: u32, gravity_w: u32, ml_w: u32 }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct MLScoreUpdated { ml_score: u32, ml_score_timestamp: u64, ml_model_version: u32 }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct ThresholdUpdated { threshold: u32 }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct CooldownUpdated { cooldown_seconds: u32 }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct RolesUpdated { pusher_role: ContractAddress, ml_oracle: ContractAddress }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct AdminChanged { new_admin: ContractAddress }
    #[derive(Copy, Drop, Serde, starknet::Event)]
    struct TradeExecuted { trade_id: felt252, direction: u8, amount: u128, score: u32, timestamp: u64, improved_price_bps: u32 }

    #[event]
    #[derive(Drop, Serde, starknet::Event)]
    enum Event {
        PredictionUpdated: PredictionUpdated,
        WeightsUpdated: WeightsUpdated,
        MLScoreUpdated: MLScoreUpdated,
    ThresholdUpdated: ThresholdUpdated,
    CooldownUpdated: CooldownUpdated,
    RolesUpdated: RolesUpdated,
    AdminChanged: AdminChanged,
    TradeExecuted: TradeExecuted,
    }

    #[constructor]
    fn constructor(ref self: ContractState, admin: ContractAddress, pusher_role: ContractAddress) {
        self.admin.write(admin);
        self.pusher_role.write(pusher_role);
        self.ml_oracle.write(admin); // default ml oracle to admin initially
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

    #[external(v0)]
    fn get_state(self: @ContractState) -> (u32, u32, u32, u32, u32) {
        (
            self.execution_threshold.read(),
            self.cooldown_seconds.read(),
            self.formula_version.read(),
            self.normalization_version.read(),
            self.composite_score.read()
        )
    }

    #[external(v0)]
    fn get_weights(self: @ContractState) -> (u32, u32, u32) {
        (
            self.astrology_w.read(),
            self.gravity_w.read(),
            self.ml_w.read()
        )
    }

    // --- Helpers ---
    fn ensure_admin(self: @ContractState) {
        assert(get_caller_address() == self.admin.read(), 'NOT_ADMIN');
    }

    fn ensure_pusher(self: @ContractState) {
        assert(get_caller_address() == self.pusher_role.read(), 'NOT_PUSHER');
    }

    fn ensure_ml_oracle(self: @ContractState) {
        assert(get_caller_address() == self.ml_oracle.read(), 'NOT_ORACLE');
    }

    // composite helper (pure)
    pub fn compute_composite(a: u32, g: u32, ml: u32, aw: u32, gw: u32, mw: u32) -> u32 {
        let total = aw + gw + mw;
        if total == 0 { return 0; }
        (a * aw + g * gw + ml * mw) / total
    }

    #[external(v0)]
    fn set_prediction_inputs(ref self: ContractState, astrology_score: u32, gravity_score: u32) {
    ensure_pusher(@self);
        assert(astrology_score <= 100_u32, 'INVALID_SCORE');
        assert(gravity_score <= 100_u32, 'INVALID_SCORE');
        let ts: u64 = get_block_timestamp().into();
        let cd: u64 = self.cooldown_seconds.read().into();
        if cd > 0_u64 && ts < self.last_input_timestamp.read() + cd { assert(false, 'COOLDOWN'); }
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
    self.emit(Event::PredictionUpdated(PredictionUpdated { astrology: astrology_score, gravity: gravity_score, ml: self.ml_score.read(), composite, formula_version: self.formula_version.read(), normalization_version: self.normalization_version.read() }));
    }

    #[external(v0)]
    fn set_ml_score(ref self: ContractState, ml_score: u32, ml_model_version: u32) {
    ensure_ml_oracle(@self);
        assert(ml_score <= 100_u32, 'INVALID_SCORE');
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
    self.emit(Event::MLScoreUpdated(MLScoreUpdated { ml_score, ml_score_timestamp: self.ml_score_timestamp.read(), ml_model_version }));
    }

    #[external(v0)]
    fn set_weights(ref self: ContractState, astrology_w: u32, gravity_w: u32, ml_w: u32) {
    ensure_admin(@self);
        assert(astrology_w + gravity_w + ml_w > 0_u32, 'ZERO_WEIGHTS');
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
    self.emit(Event::WeightsUpdated(WeightsUpdated { astrology_w, gravity_w, ml_w }));
    }

    // --- Admin / Access Management Functions ---

    #[external(v0)]
    fn update_threshold(ref self: ContractState, threshold: u32) {
    ensure_admin(@self);
        self.execution_threshold.write(threshold);
        self.emit(Event::ThresholdUpdated(ThresholdUpdated { threshold }));
    }

    #[external(v0)]
    fn update_cooldown(ref self: ContractState, cooldown_seconds: u32) {
    ensure_admin(@self);
        self.cooldown_seconds.write(cooldown_seconds);
        self.emit(Event::CooldownUpdated(CooldownUpdated { cooldown_seconds }));
    }

    #[external(v0)]
    fn set_roles(ref self: ContractState, pusher_role: ContractAddress, ml_oracle: ContractAddress) {
    ensure_admin(@self);
        self.pusher_role.write(pusher_role);
        self.ml_oracle.write(ml_oracle);
        self.emit(Event::RolesUpdated(RolesUpdated { pusher_role, ml_oracle }));
    }

    #[external(v0)]
    fn change_admin(ref self: ContractState, new_admin: ContractAddress) {
    ensure_admin(@self);
        self.admin.write(new_admin);
        self.emit(Event::AdminChanged(AdminChanged { new_admin }));
    }

    #[external(v0)]
    fn execute_trade(ref self: ContractState, trade_id: felt252, amount: u128, direction: u8, improved_price_bps: u32) {
        // Basic direction validation
        assert(direction == 0_u8 || direction == 1_u8, 'BAD_DIRECTION');
        // Threshold check
        let composite = self.composite_score.read();
        assert(composite >= self.execution_threshold.read(), 'LOW_SCORE');
        // Emit event only (event-sourced trade log strategy)
        let ts: u64 = get_block_timestamp().into();
        self.emit(Event::TradeExecuted(TradeExecuted { trade_id, direction, amount, score: composite, timestamp: ts, improved_price_bps }));
    }
}
