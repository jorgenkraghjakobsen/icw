# IC Design Flow for CP3/CP4 Projects

## Overview

This document describes the IC design flow used for the CP3 and CP4 tape-out projects on TSMC 180nm technology. The flow combines commercial Cadence tools for analog/mixed-signal design with open-source digital tools for a complete IC implementation.

## Technology

- **Process**: TSMC 180nm (G180)
- **CP3 Status**: Full mask tape-out completed
- **CP4 Status**: In preparation - addressing CP3 issues

## Design Flow Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   IC Design Flow                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐      ┌──────────────────┐        │
│  │  Analog/Mixed    │      │   Digital RTL    │        │
│  │   Signal Design  │      │     Design       │        │
│  │   (Cadence)      │      │  (Verilog/VHDL)  │        │
│  └────────┬─────────┘      └────────┬─────────┘        │
│           │                         │                   │
│           v                         v                   │
│  ┌──────────────────┐      ┌──────────────────┐        │
│  │  Schematic       │      │   Synthesis      │        │
│  │  Simulation      │      │   (OpenROAD)     │        │
│  │  (Cadence)       │      │                  │        │
│  └────────┬─────────┘      └────────┬─────────┘        │
│           │                         │                   │
│           v                         v                   │
│  ┌──────────────────┐      ┌──────────────────┐        │
│  │  Layout Design   │      │  Place & Route   │        │
│  │  (Cadence)       │      │  (OpenROAD)      │        │
│  └────────┬─────────┘      └────────┬─────────┘        │
│           │                         │                   │
│           └────────┬────────────────┘                   │
│                    v                                    │
│           ┌──────────────────┐                          │
│           │  Full Chip       │                          │
│           │  Integration     │                          │
│           │  & Verification  │                          │
│           └──────────────────┘                          │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## Tool Stack

### Analog/Mixed-Signal Path (Cadence)

**Tools Used**:
- Cadence Virtuoso for schematic capture
- Cadence Spectre for circuit simulation
- Cadence Layout for custom analog layout
- Cadence Assura/PVS for DRC/LVS verification

**Design Types**:
- Analog IP blocks (bias circuits, bandgaps, references, etc.)
- Mixed-signal blocks (ADCs, DACs, PLLs, etc.)
- Custom analog layouts
- Process-specific analog components

### Digital Path (Open Source)

**Core Tools**:
- **OpenROAD**: Complete RTL-to-GDSII flow
  - Synthesis (Yosys integration)
  - Floorplanning
  - Placement (RePlAce)
  - Clock Tree Synthesis (TritonCTS)
  - Global Routing (FastRoute)
  - Detailed Routing (TritonRoute)
  - Timing analysis (OpenSTA)

- **OpenROAD Flow Scripts (ORFS)**:
  - Automated flow management
  - Platform-specific configurations
  - Design rule checking
  - Build system integration

**Standard Cell Library**:
- TSMC G180 standard cell kit
- Custom platform adaptation for OpenROAD
- Characterized for 180nm process

## Standard Cell Platform

### TSMC G180 Platform Adaptation

The platform provides:
- Liberty timing files (.lib)
- LEF physical layout abstracts
- Technology LEF (layer definitions)
- Design rules for 180nm process
- Standard cell library optimized for OpenROAD

**Platform Location**: Adapted TSMC G180 platform for OpenROAD Flow Scripts

## Component Organization

Components are managed using the ICW workspace system:

### Analog Components
- Stored in SVN under `analog/`
- Cadence design libraries
- Custom layouts and schematics
- Example: `analog/bias`, `analog/bandgap_1v2`

### Digital Components
- Stored in SVN under `digital/`
- RTL source files (Verilog, SystemVerilog, VHDL)
- Testbenches and verification
- Example: `digital/spi_master`, `digital/top`

### Setup Components
- Stored in SVN under `setup/`
- Build scripts and configuration
- Platform-specific settings

### Process Components
- Stored in SVN under `process_setup/`
- PDK files and technology data
- Design rules and extraction decks

## Dependency Management

Components declare dependencies in `depend.config` files:

```perl
# Example: digital/top/depend.config
use component("digital/spi_master", "digital", "trunk")
use component("analog/bias", "analog", "tags/v1.0")
use component("setup/digital_env", "setup")
```

ICW automatically resolves and checks out all dependencies recursively.

## Build Flow

### Digital Build Process

1. **Workspace Setup**:
   ```bash
   icw update  # Checkout all components and dependencies
   ```

2. **Generate Dependency Lists**:
   ```bash
   icw depend-ng        # Generate source file lists
   icw tree            # Verify dependency structure
   icw hdl             # View HDL files to be compiled
   ```

3. **Synthesis & Place-and-Route** (using OpenROAD Flow Scripts):
   ```bash
   # Platform-specific build using ORFS
   make DESIGN_CONFIG=./designs/your_design/config.mk
   ```

4. **Verification**:
   - DRC checks via OpenROAD
   - LVS verification
   - Timing closure (OpenSTA)
   - Power analysis

### Analog Build Process

1. **Workspace Setup**:
   ```bash
   icw update  # Checkout analog components
   ```

2. **Cadence Design Flow**:
   - Schematic design in Virtuoso
   - Circuit simulation with Spectre
   - Custom layout design
   - DRC/LVS verification with Assura/PVS

## Integration and Tape-Out

### Full Chip Assembly

1. Digital blocks delivered as hardened macros (GDS)
2. Analog blocks as custom layouts (GDS)
3. Top-level integration and floorplanning
4. Final verification:
   - Full chip DRC
   - Full chip LVS
   - Signal integrity checks
   - Power/ground network verification

### Known Issues (CP3 → CP4)

CP3 tape-out is complete but has identified issues that need addressing in CP4:
- *Issues to be documented as they are addressed*
- Focus for CP4: Build system improvements and flow optimization

## Version Control Strategy

### Component Versioning

- **Development**: Use `trunk` for active development
- **Stable**: Use `tags/vX.Y.Z` for released versions
- **Features**: Use `branches/feature_name` for experimental work

### Release Process

```bash
icw release -t v2.0 -m "Release for CP4 tape-out"
```

This creates consistent tags across all dependent components.

## File Organization

```
workspace/
├── workspace.config          # Top-level component definitions
├── analog/                   # Analog IP blocks
│   ├── bias/
│   ├── bandgap_1v2/
│   └── opamp_folded/
├── digital/                  # Digital RTL components
│   ├── top/
│   ├── spi_master/
│   └── uart/
├── setup/                    # Build and setup scripts
│   └── digital_env/
└── process_setup/           # PDK and process files
    └── tsmc180/
```

## Benefits of This Flow

### Analog Path (Cadence)
- Industry-proven tools for analog design
- Accurate simulation and verification
- Extensive PDK support from foundry
- Full custom layout capabilities

### Digital Path (OpenROAD)
- **Open Source**: No licensing costs
- **Transparent**: Full visibility into algorithms
- **Modern**: Active development and improvements
- **Flexible**: Easy to customize and extend
- **Community**: Large user base and support

### Combined Flow
- Best of both worlds: proven analog tools + modern digital flow
- Cost-effective: Commercial tools only where needed
- Reproducible: Open source digital flow
- Version controlled: All components tracked in SVN

## Future Improvements

Areas for enhancement in CP4:
- Build system automation improvements
- Better integration between analog and digital domains
- Enhanced verification methodologies
- Streamlined tape-out preparation
- Issue resolution from CP3 learnings

## References

- **OpenROAD**: https://theopenroadproject.org/
- **OpenROAD Flow Scripts**: https://github.com/The-OpenROAD-Project/OpenROAD-flow-scripts
- **TSMC 180nm**: Contact TSMC for PDK documentation
- **ICW Documentation**: See `TREE_HDL_COMMANDS.md`, `DEPENDENCY_TRACKING.md`

## Contacts and Resources

- **Build System**: ICW workspace management
- **Analog Support**: Cadence support and internal team
- **Digital Support**: OpenROAD community and documentation
- **PDK**: TSMC support and FAE team
