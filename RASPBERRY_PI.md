# Raspberry Pi Deployment Guide

This guide covers optimizations and monitoring tools for running Halko on resource-constrained hardware like Raspberry Pi 3B (1GB RAM).

## Memory-Optimized Builds

**Essential for 1GB RAM systems.** Use the `OPTIMIZED=yes` flag to reduce memory footprint by ~30%:

```bash
# Build optimized binaries (reduces memory footprint by ~30%)
OPTIMIZED=yes make build

# Install to system
sudo make install
sudo make install-webapp
sudo make systemd-units
```

### Optimization Flags

The optimized builds are controlled via the `OPTIMIZED` variable in the Makefile. Both standard and optimized flag sets are defined at the top of the Makefile - you can either:

1. **Use the variable**: `OPTIMIZED=yes make build`
2. **Edit the Makefile**: Comment/uncomment the flag definitions to change the default

**Optimized flags applied when `OPTIMIZED=yes`:**

| Flag | Purpose | Impact |
|------|---------|--------|
| `-ldflags="-s"` | Strip debug information | ~20-25% size reduction |
| `-ldflags="-w"` | Strip DWARF symbol table | ~5-10% size reduction |
| `-trimpath` | Remove absolute file paths | Smaller binaries, reproducible builds |

**Trade-offs:**

- ✅ **Smaller binaries**: 30-40% reduction in disk and memory footprint
- ✅ **Lower RAM usage**: Reduced runtime memory consumption
- ✅ **Faster startup**: Less data to load into memory
- ⚠️ **No debugging symbols**: Cannot use `gdb` or `dlv` debugger on binaries
- ⚠️ **Slightly slower**: ~5-10% performance reduction (negligible for this application)

### Binary Size Comparison

Typical binary sizes on ARM (Raspberry Pi 3B):

| Service | Standard | Optimized | Savings |
|---------|----------|-----------|---------|
| controlunit | ~15 MB | ~10 MB | 33% |
| powerunit | ~12 MB | ~8 MB | 33% |
| sensorunit | ~11 MB | ~7.5 MB | 32% |
| halkoctl | ~13 MB | ~9 MB | 31% |

## Memory Monitoring

For long-running tests on resource-constrained hardware, use `monitor-memory.py` to track process memory usage and detect potential leaks:

```bash
# Basic monitoring (console output only)
./scripts/monitor-memory.py

# Save detailed log for analysis
./scripts/monitor-memory.py -o memory-test.csv -t 7200  # 2 hours
```

See `./scripts/monitor-memory.py --help` for all options.

## Raspberry Pi Hardware Recommendations

### Production Tested Configuration

**Currently running in production:**

- **Raspberry Pi 3 Model B** (1GB RAM)
- **Storage**: USB 2.0 SSD with **USB boot** (no SD card - boots directly from SSD)
- **Power**: Official 5V 2.5A power supply
- **Cooling**: Heatsink or passive cooling sufficient for continuous operation

**Performance notes:**

- Memory usage stays comfortably within 1GB with optimized builds (~150 MB peak)
- USB boot provides excellent I/O performance and eliminates SD card failure risks
- CPU typically <30% during active program execution
- 24/7 uptime reliability with USB SSD boot

### Minimum Requirements

- **Raspberry Pi 3 Model B** (1GB RAM) - proven adequate
- **Storage**: USB SSD with USB boot enabled (no SD card needed)
  - Requires bootloader update (see USB Boot Setup below)
  - Eliminates SD card wear and failure issues
  - Better I/O performance than SD cards
- **Power**: Official power supply for model (3B: 2.5A, 4B: 3A)
- **Cooling**: Heatsink recommended for continuous operation

### Upgrade Path

If you need more headroom:

- **Raspberry Pi 4 Model B** (2GB+ RAM) - faster CPU, USB 3.0, dual display, native USB boot
- **Raspberry Pi 5** (4GB+ RAM) - significant performance improvement, native USB boot

**Note**: Pi 4 and Pi 5 have native USB boot support (no bootloader update needed)

### Performance Expectations

Memory usage during normal operation (systemd deployment):

| Service | Idle | Running Program |
|---------|------|-----------------|
| controlunit | 30-40 MB | 40-55 MB |
| powerunit | 25-35 MB | 30-40 MB |
| sensorunit | 20-30 MB | 25-35 MB |
| nginx (webapp) | 5-10 MB | 10-15 MB |
| **Total** | **~90 MB** | **~150 MB** |

**Note**: Total usage stays well under 200 MB, leaving plenty of headroom on 1GB RAM systems.

## Raspberry Pi OS Setup

### USB Boot Setup (Raspberry Pi 3B)

**For USB boot on Pi 3B, you must first enable it:**

1. **One-time bootloader update** (requires temporary SD card):

   ```bash
   # Flash Raspberry Pi OS to SD card
   # Boot from SD card, then run:
   echo program_usb_boot_mode=1 | sudo tee -a /boot/config.txt
   sudo reboot
   # After reboot, verify:
   vcgencmd otp_dump | grep 17:
   # Should show: 17:3020000a (USB boot enabled)
   ```

2. **Flash Raspberry Pi OS to USB SSD**:
   - Use Raspberry Pi Imager to write OS to USB SSD
   - Select "Raspberry Pi OS Lite (32-bit)" for Pi 3B
   - Remove SD card, connect USB SSD, power on
   - Pi will now boot directly from USB

### Recommended OS

Use **Raspberry Pi OS Lite (32-bit)** for Raspberry Pi 3B, flashed to USB SSD:

```bash
# After booting from USB SSD
sudo apt update && sudo apt upgrade -y
sudo apt install -y git golang
```

### Network Setup

The Raspberry Pi uses a dual-interface network configuration:

- **WiFi (wlan0)**: Internet access and sensor unit display IP
- **Ethernet (eth0)**: Direct connection to Shelly device

#### WiFi Configuration

Configure WiFi for internet access and remote connectivity:

```bash
# Edit network configuration
sudo nano /etc/wpa_supplicant/wpa_supplicant.conf
```

Add your WiFi credentials:

```conf
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=US

network={
    ssid="YourWiFiSSID"
    psk="YourWiFiPassword"
    key_mgmt=WPA-PSK
}
```

Restart networking:

```bash
sudo systemctl restart dhcpcd
# Or reboot:
sudo reboot
```

Verify WiFi connection:

```bash
ip addr show wlan0
# Should show assigned IP address
```

#### Ethernet Configuration (Direct Shelly Connection)

Configure static IP for direct Ethernet connection to Shelly device:

```bash
sudo nano /etc/dhcpcd.conf
```

Add at the end of the file:

```conf
# Static IP for Ethernet (direct Shelly connection)
interface eth0
static ip_address=192.168.10.1/24
static domain_name_servers=192.168.10.1
```

**Configure Shelly device:**

1. Connect to Shelly via its WiFi AP or web interface
2. Set Shelly to use static IP: `192.168.10.2`
3. Set netmask: `255.255.255.0`
4. Set gateway: `192.168.10.1` (the Raspberry Pi)
5. Connect Shelly directly to Pi's Ethernet port

Restart networking:

```bash
sudo systemctl restart dhcpcd
```

Verify Ethernet configuration:

```bash
ip addr show eth0
# Should show: 192.168.10.1/24

# Test Shelly connectivity
ping -c 3 192.168.10.2
```

#### Remote Access (OpenVPN "Call Home")

For remote access to deployed Raspberry Pi systems, configure OpenVPN client to connect to your VPN server:

**Prerequisites:**

- OpenVPN client configuration file from your VPN server (e.g., `client.ovpn`)
- VPN server that accepts client connections

**Install OpenVPN:**

```bash
sudo apt update
sudo apt install -y openvpn
```

**Configure OpenVPN client:**

```bash
# Copy your client configuration file to OpenVPN directory
# (Rename to .conf extension - required for systemd service)
sudo cp /path/to/your/client.ovpn /etc/openvpn/client/halko-vpn.conf

# If your config references separate certificate/key files, copy them too:
# sudo cp ca.crt client.crt client.key /etc/openvpn/client/
```

**Enable OpenVPN to start on boot:**

```bash
# Enable the OpenVPN client service
sudo systemctl enable openvpn-client@halko-vpn

# Start the service
sudo systemctl start openvpn-client@halko-vpn

# Check status
sudo systemctl status openvpn-client@halko-vpn
```

**Verify connection:**

```bash
# Check if VPN interface is up
ip addr show tun0

# Check OpenVPN logs
sudo journalctl -u openvpn-client@halko-vpn -f

# Test connectivity through VPN
# (Replace with your VPN server's internal IP or test endpoint)
ping -c 3 10.8.0.1
```

**Configuration notes:**

- The service name must match the config filename: `halko-vpn.conf` → `openvpn-client@halko-vpn`
- OpenVPN will automatically reconnect on network failures or reboots
- VPN connection typically establishes after network interfaces are up (WiFi/Ethernet)
- If using authentication files, ensure they're readable by root: `sudo chmod 600 /etc/openvpn/client/*`

**Troubleshooting:**

```bash
# View detailed logs
sudo journalctl -u openvpn-client@halko-vpn --no-pager

# Restart service
sudo systemctl restart openvpn-client@halko-vpn

# Disable autostart (if needed)
sudo systemctl disable openvpn-client@halko-vpn
```

### Install Halko

```bash
cd ~
git clone https://github.com/rmkhl/halko.git
cd halko
make prepare
OPTIMIZED=yes make build
sudo make install
sudo make install-webapp
sudo make systemd-units
```

### Configure Before Starting

Edit `/etc/opt/halko.cfg` to match your hardware:

```bash
sudo nano /etc/opt/halko.cfg
```

**Critical settings:**

- `network_interface`: Set to `wlan0` (WiFi interface)
  - This IP address will be displayed on the sensor unit for connection purposes
  - Allows remote access to webapp and API from network devices
- `serial_device`: Arduino path (e.g., `/dev/ttyUSB0` or `/dev/ttyACM0`)
  - Run `ls /dev/ttyUSB* /dev/ttyACM*` to find Arduino device
- `shelly_address`: `192.168.10.2` (Shelly static IP on Ethernet)

**Example configuration:**

```json
{
  "controlunit": {
    "network_interface": "wlan0",
    "api_port": 8090,
    ...
  },
  "api_endpoints": {
    "powerunit": "http://192.168.10.2/..."
  },
  ...
}
```

See [templates/README.md](templates/README.md) for full configuration details.

### Start Services

```bash
# Services are already enabled and started by systemd-units target
# Check status:
sudo systemctl status halko@controlunit
sudo systemctl status halko@powerunit
sudo systemctl status halko@sensorunit

# View logs:
sudo journalctl -u halko@controlunit -f
```

### Access WebApp

Navigate to `http://raspberry-pi-ip/` in your browser.

## Troubleshooting

### Out of Memory Errors

If you encounter OOM kills:

1. **Check available memory**:

   ```bash
   free -h
   htop
   ```

2. **Monitor service memory usage**:

   ```bash
   # Real-time monitoring
   ./scripts/monitor-memory.py -p controlunit powerunit sensorunit

   # Or check individual service
   ps aux | grep controlunit
   ```

3. **Add swap file if needed** (1GB):

   ```bash
   sudo fallocate -l 1G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   # Make permanent:
   echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
   ```

4. **Reduce log verbosity**:
   pass `-loglevel 1` instead of `-loglevel 2` in systemd units

### Build Failures

If compilation fails due to low memory:

```bash
# Limit concurrent builds
GOMAXPROCS=1 make build

# Or build with optimization flags
GOMAXPROCS=1 OPTIMIZED=yes make build

# Or build services individually
go build -ldflags="-s -w" -trimpath -o bin/controlunit ./controlunit/
```

### Slow Performance

If services are sluggish:

1. **Check CPU throttling**:

   ```bash
   vcgencmd measure_temp
   vcgencmd get_throttled
   ```

   If throttled (0x50000 or higher), improve cooling

2. **Reduce log verbosity** - edit systemd units to use `-loglevel 1`

3. **Monitor resource usage**:

   ```bash
   htop
   ./scripts/monitor-memory.py
   ```

### Network Connectivity Issues

If services cannot communicate:

1. **Verify WiFi connection**:

   ```bash
   ip addr show wlan0
   iwconfig wlan0
   ping -c 3 8.8.8.8  # Test internet
   ```

2. **Verify Ethernet connection to Shelly**:

   ```bash
   ip addr show eth0  # Should show: 192.168.10.1/24
   ping -c 3 192.168.10.2  # Test Shelly connectivity
   ```

3. **Check Shelly is accessible**:

   ```bash
   curl http://192.168.10.2/status
   # Should return Shelly status JSON
   ```

4. **Verify controlunit shows correct IP on sensor display**:

   ```bash
   # Check what IP controlunit is broadcasting:
   sudo journalctl -u halko@controlunit | grep -i "heartbeat\|ip\|interface"

   # Verify WiFi IP:
   hostname -I
   ```

5. **Check firewall isn't blocking connections**:

   ```bash
   sudo iptables -L
   # If firewall is active, may need to allow ports 8090-8093
   ```

6. **Restart networking if needed**:

   ```bash
   sudo systemctl restart dhcpcd
   sudo systemctl restart wpa_supplicant
   ```

## Production Deployment Tips

1. **Use optimized builds** for better performance:

   ```bash
   OPTIMIZED=yes make build
   ```

2. **Services auto-restart on crash** (already configured in systemd units):

   ```ini
   [Service]
   Restart=always
   RestartSec=10
   ```

3. **Log rotation** for disk space management:

   ```bash
   sudo journalctl --vacuum-time=7d
   sudo journalctl --vacuum-size=100M
   ```

4. **Monitor disk usage** on USB SSD:

   ```bash
   df -h /
   du -sh /var/opt/halko/*
   ```

5. **Backup configuration** regularly:

    ```bash
    # Backup config and programs
    sudo tar czf halko-backup-$(date +%Y%m%d).tar.gz \
    /etc/opt/halko.cfg \
    /var/opt/halko/programs
    ```

## See Also

- [README.md](README.md) - Main project documentation
- [API.md](API.md) - REST API reference
- [PROGRAM.md](PROGRAM.md) - Program structure and validation
