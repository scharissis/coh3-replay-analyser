# Command Parsing TODO List

  SCMD = Simulation Command (or Synchronized Command)
  DCMD = Display Command (or Desynchronized Command)

  SCMD Commands (Simulation/Synchronized)
  These affect the actual game state and must be synchronized across all players:
  - SCMD_Move - Unit movement
  - SCMD_Attack - Combat actions
  - SCMD_Capture - Capturing strategic points
  - SCMD_BuildStructure - Construction
  - SCMD_Retreat - Unit retreats
  - SCMD_Upgrade - Unit/building upgrades
  - SCMD_Reinforce - Adding troops to squads

## Command Types Analysis (from vault-debug-2_07_2025__12_22_PM.txt)

### Unparsed Commands ❌
These commands are not parsed and need implementation:

#### High Priority (Most Frequent)
1. **DCMD_CameraTrack** (8,362 occurrences)
   - Currently: Unknown command with action_type
   - Priority: High
   - Status: ❌ Not parsed

2. **DCMD_COUNT** (3,198 occurrences)
   - Currently: Unknown command with action_type
   - Priority: High
   - Status: ❌ Not parsed

3. **SCMD_Move** (2,548 occurrences)
   - Currently: Unknown command with action_type
   - Priority: High
   - Status: ❌ Not parsed

#### Medium Priority
4. **SCMD_Attack** (234 occurrences)
   - Currently: Unknown command with action_type
   - Priority: Medium
   - Status: ❌ Not parsed

5. **SCMD_Capture** (154 occurrences)
   - Currently: Unknown command with action_type
   - Priority: Medium
   - Status: ❌ Not parsed

#### Lower Priority (Less Frequent)
6. **SCMD_Stop** (94 occurrences)
7. **SCMD_BuildStructure** (78 occurrences)
8. **SCMD_Reinforce** (62 occurrences)
9. **SCMD_SetDefaultAction** (52 occurrences)
10. **SCMD_Retreat** (39 occurrences)
11. **SCMD_Upgrade** (36 occurrences)
12. **SCMD_SetStance** (35 occurrences)
13. **SCMD_InstantSetupBag** (26 occurrences)
14. **SCMD_SetRallyPoint** (23 occurrences)
15. **SCMD_BuildSquad** (14 occurrences)
16. **SCMD_Unload** (10 occurrences)
17. **SCMD_Load** (9 occurrences)
18. **SCMD_Destroy** (8 occurrences)
19. **SCMD_Face** (5 occurrences)
20. **SCMD_Ability** (4 occurrences)
21. **SCMD_Salvage** (4 occurrences)
22. **SCMD_ReinforceUnit** (3 occurrences)
23. **SCMD_InstantReinforce** (2 occurrences)
24. **SCMD_Camouflage** (2 occurrences)
25. **SCMD_SetAutoTargetType** (1 occurrence)
26. **SCMD_CancelProduction** (1 occurrence)

## Implementation Tasks

### Phase 1: High Priority Commands
- [ ] Implement DCMD_CameraTrack parser
- [ ] Implement DCMD_COUNT parser
- [ ] Implement SCMD_Move parser
- [ ] Add pbgid lookup support for movement commands

### Phase 2: Medium Priority Commands
- [ ] Implement SCMD_Attack parser
- [ ] Implement SCMD_Capture parser
- [ ] Add pbgid lookup support for attack/capture commands

### Phase 3: Lower Priority Commands
- [ ] Implement SCMD_Stop parser
- [ ] Implement SCMD_BuildStructure parser
- [ ] Implement SCMD_Reinforce parser
- [ ] Implement SCMD_SetDefaultAction parser
- [ ] Implement SCMD_Retreat parser
- [ ] Implement SCMD_Upgrade parser
- [ ] Implement SCMD_SetStance parser
- [ ] Implement remaining SCMD_* parsers

### Phase 4: Research and Enhancement
- [ ] Research command structure patterns to identify common pbgid patterns
- [ ] Implement pbgid lookup hydration for all new command types
- [ ] Add unit tests for new command parsers
- [ ] Update documentation with new command types

## Previous TODOs
- [ ] Change CLI to https://github.com/urfave/cli
- [ ] Additional validation: Test edge cases, different replay types

