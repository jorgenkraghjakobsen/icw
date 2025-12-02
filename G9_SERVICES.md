# G9 Server Services Documentation

**Server**: g9 (anyvej11.dk / icworks.dk)
**IP Address**: 87.104.43.154 (public), 192.168.1.31 (local)
**Last Updated**: 2025-12-01

---

## Overview

The g9 server hosts multiple web services and development tools for IC design workflows, SVN repository management, and internal tooling. All services are accessible on the local network, with some exposed through nginx reverse proxy.

---

## Network-Accessible Services

### Port 80/443 - HTTP/HTTPS (Nginx)
- **Service**: Nginx web server and reverse proxy
- **Status**: ✅ Active
- **Binding**: All interfaces (0.0.0.0)
- **Access**:
  - http://anyvej11.dk
  - http://icworks.dk
  - https://anyvej11.dk
  - https://icworks.dk
- **Purpose**: Main web server, reverse proxy for backend services
- **Systemd**: `nginx.service`

### Port 3000 - ICView Viewport
- **Service**: ICView - IC Layout Viewer
- **Status**: ✅ Active
- **Binding**: All interfaces (*:3000)
- **Access**:
  - http://anyvej11.dk:3000
  - http://192.168.1.31:3000
- **Title**: "ICView Viewport"
- **Purpose**: Web-based IC layout viewing and analysis tool
- **Location**: `/home/mihme/work/ICView/`
- **Technology**: Vite (JavaScript framework)

### Port 5000 - Node.js Service
- **Service**: Node.js application
- **Status**: ⚠️ Running but returning errors
- **Binding**: All interfaces (*:5000)
- **Access**: http://192.168.1.31:5000
- **Process**: node (PID: 442262)
- **Purpose**: Unknown/Development service

### Port 7888 - Web Application
- **Service**: Unknown web service
- **Status**: ✅ Active
- **Binding**: All interfaces (*:7888)
- **Access**: http://192.168.1.31:7888
- **Title**: "Home"
- **Purpose**: Unknown application (requires investigation)

### Port 8080 - IC Works Network Monitor
- **Service**: Network monitoring dashboard
- **Status**: ✅ Active
- **Binding**: All interfaces (*:8080)
- **Access**:
  - http://anyvej11.dk:8080
  - http://192.168.1.31:8080
- **Title**: "IC Works Network Monitor"
- **Purpose**: Monitor network topology and service status
- **Location**: `/home/mihme/work/icw-dashboard/`
- **Features**:
  - Real-time network topology visualization
  - Service status monitoring
  - Device status tracking
  - Uses Vis.js for network visualization

### Port 8081 - ICW Repository Manager ⭐
- **Service**: ICW SVN Repository Management Interface
- **Status**: ✅ Active
- **Binding**: All interfaces (*:8081)
- **Access**:
  - http://anyvej11.dk:8081/icw/
  - http://192.168.1.31:8081/icw/
- **Title**: "Go Index"
- **Purpose**: Web interface for SVN repository management
- **Location**: `/home/mihme/work/icw-maw/` (running service)
- **Systemd**: `icw-maw.service`
- **Technology**: Go web service
- **Endpoints**:
  - `/icw/list_repos` - List all SVN repositories
  - `/icw/list_users` - List and manage users
  - `/icw/new_repo` - Create new repository
  - `/icw/repo?repo=NAME` - View specific repository
- **Available Repositories**:
  - Marcus
  - cp2
  - cp3 (primary active repository)
  - icworks
  - paral
  - sensi
  - svn_testing
  - testrepo

### Port 8082 - Register Database Interface
- **Service**: Register/Memory Map Management
- **Status**: ✅ Active
- **Binding**: All interfaces (*:8082)
- **Access**: http://192.168.1.31:8082
- **Title**: "Home"
- **Purpose**: Register and section database interface
- **Location**: `/home/mihme/work/reg-db-interface/`
- **Technology**: Go web service with Bootstrap UI
- **Processes**:
  - Multiple instances running (PIDs: 2706, 2708, 2715, 2720)

### Port 8083 - Christmas Series
- **Service**: Media/Entertainment service
- **Status**: ✅ Active
- **Binding**: All interfaces (0.0.0.0:8083, [::]:8083)
- **Access**: http://192.168.1.31:8083
- **Title**: "Christmas Series - 24 Episodes"
- **Purpose**: Media streaming/content service

### Port 8084 - Registers and Sections
- **Service**: Register documentation/management
- **Status**: ✅ Active
- **Binding**: All interfaces (*:8084)
- **Access**: http://192.168.1.31:8084
- **Title**: "Registers and Sections"
- **Purpose**: Hardware register documentation and management interface

---

## Development Services (Localhost Only)

### Port 3306 - MySQL/MariaDB
- **Service**: Database server
- **Status**: ✅ Active
- **Binding**: Localhost only (127.0.0.1:3306)
- **Access**: Local only
- **Purpose**: Backend database for web applications

### Port 6379 - Redis
- **Service**: Redis key-value store
- **Status**: ✅ Active
- **Binding**: Localhost only (127.0.0.1:6379, [::1]:6379)
- **Access**: Local only
- **Systemd**: `redis-server.service`
- **Purpose**: Cache and session storage for web applications

### Port 631 - CUPS
- **Service**: Print service
- **Status**: ✅ Active
- **Binding**: Localhost only (127.0.0.1:631, [::1]:631)
- **Access**: Local only
- **Purpose**: Print server

### Ports 6010-6015 - X11 Forwarding
- **Service**: SSH X11 forwarding
- **Status**: ✅ Active
- **Binding**: Localhost only
- **Access**: Local only
- **Purpose**: X11 display forwarding for SSH sessions

### Port 33165 - Unknown Service
- **Service**: Unknown
- **Status**: ✅ Active
- **Binding**: Localhost only (127.0.0.1:33165)
- **Purpose**: Unknown (requires investigation)

---

## Infrastructure Services

### Port 22 - SSH
- **Service**: OpenSSH server
- **Status**: ✅ Active
- **Binding**: All interfaces (0.0.0.0:22, [::]:22)
- **Access**: ssh://anyvej11.dk:22
- **Purpose**: Remote administration and access

### Port 53 - DNS
- **Service**: DNS server (systemd-resolved + libvirt)
- **Status**: ✅ Active
- **Binding**:
  - 127.0.0.53:53 (systemd-resolved)
  - 192.168.122.1:53 (libvirt network)
- **Purpose**: Local DNS resolution and VM networking

### Port 139/445 - Samba (SMB/CIFS)
- **Service**: Samba file sharing
- **Status**: ✅ Active
- **Binding**: All interfaces (0.0.0.0:139, 0.0.0.0:445)
- **Purpose**: Windows file sharing

### Port 3690 - Subversion (SVN) ⭐
- **Service**: SVN server (svnserve)
- **Status**: ✅ Active
- **Binding**: All interfaces (0.0.0.0:3690)
- **Access**: svn://anyvej11.dk
- **Systemd**: `svnserve.service`
- **Repository Root**: `/data_v1/svn/repos/`
- **Purpose**: Version control for IC design components
- **Authentication**: SASL (username/password)
- **Process**: Running as root (PID: 1177)
- **Command**: `/usr/bin/svnserve -d --foreground -r /data_v1/svn/repos`

### Port 24678 - Unknown Service
- **Service**: Unknown
- **Status**: ✅ Active
- **Binding**: All interfaces (*:24678)
- **Purpose**: Unknown (requires investigation)

---

## SVN Repository Structure

Located at `/data_v1/svn/repos/`, accessible via `svn://anyvej11.dk/`

### Active Repositories:
- **cp3** - Primary active repository
  - `components/analog/` - Analog/mixed-signal designs
  - `components/digital/` - Digital HDL designs
  - `components/setup/analog/` - Analog toolchain setup
  - `components/setup/digital/` - Digital toolchain setup
  - `components/setup/flow_platform_tsmc_tcb018gbwp7t/` - TSMC 180nm PDK

- **cp2** - Legacy repository
- **Marcus** - User repository
- **icworks** - Empty/unused
- **paral** - Project repository
- **sensi** - Project repository
- **svn_testing** - Test repository
- **testrepo** - Test repository

### Deleted/Archived:
Located at `/data_v1/svn/deleted/`:
- certus
- Marcusa
- n

---

## Management Tools

### Command Line (Makefile)
Located: `/home/jakobsen/work/icw_backend/makefile`

**Repository Management:**
```bash
make repo_add REPO=name        # Create new repository
make repo_del REPO=name        # Delete repository (disabled)
```

**User Management:**
```bash
make user_add REPO=name USER=username    # Add user to repo
make user_del REPO=name USER=username    # Remove user from repo
make user_list REPO=name                 # List users for repo
make user_list_all                       # List all users
```

**Service Management:**
```bash
make svn_restart              # Restart SVN server
make svn_start                # Start SVN server
```

### Web Interface
**Primary Interface:** http://anyvej11.dk:8081/icw/
- Create repositories via web form
- List and manage repositories
- User management (in development)

---

## Network Access Summary

| Port | Service | Network | localhost | Purpose |
|------|---------|---------|-----------|---------|
| 22 | SSH | ✅ | ✅ | Remote access |
| 53 | DNS | Local subnet | ✅ | DNS resolution |
| 80 | HTTP (nginx) | ✅ | ✅ | Web server |
| 139/445 | Samba | ✅ | ✅ | File sharing |
| 443 | HTTPS (nginx) | ✅ | ✅ | Secure web |
| 631 | CUPS | ❌ | ✅ | Printing |
| 3000 | ICView | ✅ | ✅ | Layout viewer |
| 3306 | MySQL | ❌ | ✅ | Database |
| 3690 | SVN | ✅ | ✅ | Version control |
| 5000 | Node.js | ✅ | ✅ | Development |
| 6010-6015 | X11 | ❌ | ✅ | SSH forwarding |
| 6379 | Redis | ❌ | ✅ | Cache/sessions |
| 7888 | Web App | ✅ | ✅ | Unknown |
| 8080 | Network Monitor | ✅ | ✅ | Monitoring |
| 8081 | ICW Manager | ✅ | ✅ | SVN management |
| 8082 | Register DB | ✅ | ✅ | Register mgmt |
| 8083 | Media | ✅ | ✅ | Entertainment |
| 8084 | Register Docs | ✅ | ✅ | Documentation |
| 24678 | Unknown | ✅ | ✅ | Unknown |

---

## System Information

**Operating System**: Ubuntu 22.04 LTS (based on service versions)
**Hostname**: g9
**Domain**: anyvej11.dk, icworks.dk (aliases, same IP)
**Network Interfaces**:
- 192.168.1.31 (primary)
- 192.168.122.1 (libvirt/KVM bridge)
- 172.17.0.1 (Docker bridge)

**Key System Services**:
- `nginx.service` - Web server
- `svnserve.service` - SVN server
- `redis-server.service` - Redis cache
- `icw-maw.service` - ICW web interface

---

## Access Methods

### For ICW Users:

**Repository Access:**
```bash
export ICW_REPO=cp3
export ICW_SVN_URL=svn://anyvej11.dk
icw update
```

**Web Interfaces:**
- SVN Management: http://anyvej11.dk:8081/icw/
- Network Monitor: http://anyvej11.dk:8080
- ICView: http://anyvej11.dk:3000

### For Administrators:

**SSH Access:**
```bash
ssh g9.local
ssh jakobsen@anyvej11.dk
```

**Service Management:**
```bash
systemctl status svnserve
systemctl status nginx
systemctl status icw-maw
```

---

## Notes

1. **Port Conflicts**: Port 8080 is already in use by Network Monitor. New services should use ports 8085+ or other available ports.

2. **Security**: Most development services (MySQL, Redis) are bound to localhost only, which is good security practice.

3. **Multiple Instances**: Some services (reg-db-interface) have multiple running instances, likely for development/testing.

4. **Unmapped Services**: Ports 5000, 7888, 8083, 24678, and 33165 need further investigation to determine their exact purpose.

5. **Domain Aliases**: Both anyvej11.dk and icworks.dk point to the same server (87.104.43.154). Use anyvej11.dk as the primary domain for consistency.

6. **SVN Authentication**: Different hostnames (anyvej11.dk vs icworks.dk) use different authentication realms. Cached credentials may differ between them.

---

## Maintenance Tasks

### Regular:
- Monitor disk usage at `/data_v1/svn/repos/`
- Review and clean up unused repositories in `/data_v1/svn/deleted/`
- Update user access as needed via web interface or CLI

### Periodic:
- Backup SVN repositories
- Review and document unknown services
- Update service documentation
- Security audit of network-exposed services

---

## Contact

For service issues or access requests, contact the system administrator.

Repository managed services: http://anyvej11.dk:8081/icw/
