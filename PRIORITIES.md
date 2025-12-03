# ICW Development Priorities

## Current Focus: Core ICW Functionality

### Phase 1: Essential Commands (NOW) ✅
- [x] `icw update` - Checkout components with dependency resolution
- [x] `icw status` / `icw st` - Show workspace status
- [x] `icw tree` - Display dependency tree from config files
- [x] `icw hdl` - Display dependency tree with HDL files
- [x] `icw list` - List repository components
- [x] `icw test` - Test repository connection

**Status**: ✅ Complete

### Phase 2: Migration Functionality (NEXT PRIORITY)
- [ ] `icw migrate` - Migrate components between repositories (CP3 → CP4)
  - Repository creation via MAW integration
  - Component selection (interactive)
  - Full history vs latest version migration
  - Dependency handling
  - User migration

**Estimated effort**: 2-3 weeks
**Blocker**: Need repo migration for CP4 tape-out

**Plan**: See `REPO_MIGRATION_PLAN_V2.md`

### Phase 3: Build System Integration
- [ ] `icw depend-ng` - Generate dependency lists for build systems
  - Modelsim format
  - Design Compiler format
  - Incisive format
  - TCL format
  - List format

**Estimated effort**: 1 week
**Reason**: Critical for CP4 build flow

### Phase 4: Additional Core Commands
- [ ] `icw add` - Add component to repository
- [ ] `icw release` - Release component with dependencies
- [ ] `icw dumpdepend` / `icw dd` - Dump dependencies for tools
- [ ] `icw wipe` - Reset workspace to clean state

**Estimated effort**: 2-3 weeks
**Reason**: Nice to have, but not blocking

### Phase 5: Advanced Features
- [ ] Git support for tools components (partial exists)
- [ ] Component versioning improvements
- [ ] Workspace templates
- [ ] Better conflict resolution UI

**Estimated effort**: 3-4 weeks
**Reason**: Future enhancements

## Future: User Management (LATER)

### MVP User Management (Deferred)
When core functionality is stable:
- [ ] Self-service password reset
- [ ] User invitation system
- [ ] Basic role-based access (admin/user)
- [ ] Email service integration
- [ ] Simple admin UI

**Estimated effort**: 2-3 weeks
**When**: After core ICW is production-ready

**Plan**: See `USER_MANAGEMENT_PLAN.md`

## Timeline

```
┌─────────────────────────────────────────────────────────────┐
│ Current: December 2024                                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Week 1-3:  [====================] Phase 2: Migration         │
│            └─ CP3 → CP4 migration tool                       │
│                                                              │
│ Week 4:    [======] Phase 3: Build System                   │
│            └─ icw depend-ng implementation                   │
│                                                              │
│ Week 5-7:  [====================] Phase 4: Core Commands    │
│            └─ add, release, dumpdepend, wipe                 │
│                                                              │
│ Week 8+:   [============================] Phase 5+          │
│            └─ Advanced features & User Management           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Decision Rationale

### Why Migration First?
1. **Blocking**: CP4 tape-out depends on it
2. **User story**: Need to migrate components from CP3
3. **Integration**: Works with existing MAW system
4. **Impact**: Enables entire CP4 project

### Why User Management Later?
1. **Not blocking**: Current SASL auth works
2. **Workaround exists**: Admins can reset passwords manually
3. **Nice to have**: Improves UX but not critical
4. **Better timing**: After core tool is stable

### Why depend-ng Early?
1. **Build flow**: CP4 needs dependency lists
2. **Daily use**: Developers need this frequently
3. **Quick win**: Relatively simple to implement
4. **High value**: Integrates with OpenROAD flow

## Success Criteria

### Phase 2 (Migration) - Done When:
- ✅ Can run `icw migrate --from cp3 --to cp4`
- ✅ Successfully migrates selected components
- ✅ Copies users from cp3 to cp4
- ✅ Updates all depend.config references
- ✅ Handles dependencies automatically
- ✅ CP4 repository is ready for development

### Phase 3 (Build System) - Done When:
- ✅ `icw depend-ng` generates correct file lists
- ✅ Works with OpenROAD Flow Scripts
- ✅ Supports all needed formats
- ✅ Integrates with existing makefiles

## Current Status

**Active Work**: Planning Phase 2 (Migration)
**Blockers**: None - ready to start implementation
**Next Action**: Begin migration tool implementation

---

## Notes

- User management MVP plan documented in `USER_MANAGEMENT_PLAN.md`
- Migration plan in `REPO_MIGRATION_PLAN_V2.md`
- Design flow documented in `DESIGN_FLOW.md` (private)
- All essential commands (Phase 1) are complete and working
