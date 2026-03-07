# Halko Configuration Template

This directory contains templates used during production installation.

## halko.cfg

The main configuration file template installed to `/etc/opt/halko.cfg` by `make install`.

### Important Configuration Settings

Before starting the services, you **must** edit `/etc/opt/halko.cfg` to match your system:

#### 1. Network Interface (`controlunit.network_interface`)

**Default:** `"eth0"`

This must match your system's actual network interface name. The controlunit heartbeat service uses this to determine its IP address for display on the sensor unit and remote access.

To find your interface name:

```bash
ip addr show
```

Common interface names:

- `eth0` - Traditional Ethernet (older systems)
- `enp0s3`, `ens33`, `enp1s0` - Predictable PCI Ethernet names (modern Linux)
- `wlan0`, `wlp2s0` - WiFi interfaces
- Container environments typically use `eth0`

**Raspberry Pi Production Setup:**

For Raspberry Pi deployments with dual network interfaces (see [RASPBERRY_PI.md](../RASPBERRY_PI.md)):

- Set `network_interface: "wlan0"` (WiFi for display IP and remote access)
- Configure Ethernet (`eth0`) with static IP `192.168.10.1/24` for direct Shelly connection
- This allows the sensor unit to display the WiFi IP for remote access while keeping Shelly on a dedicated network

#### 2. Serial Device (`sensorunit.serial_device`)

**Default:** `"/dev/ttyUSB0"`

Path to your Arduino device for thermocouple readings.

To find connected USB serial devices:

```bash
ls -l /dev/ttyUSB* /dev/ttyACM*
```

Common paths:

- `/dev/ttyUSB0` - First USB serial adapter
- `/dev/ttyACM0` - Arduino using native USB
- `/dev/ttyUSB1` - Second USB device

#### 3. Shelly Address (`power_unit.shelly_address`)

**Default:** `"http://localhost:8088"` (simulator)

For production, change this to your actual Shelly smart switch IP address:

**Standard network setup:**

```json
"shelly_address": "http://192.168.1.50"
```

**Raspberry Pi production setup** (direct Ethernet connection):

```json
"shelly_address": "http://192.168.10.2"
```

See [RASPBERRY_PI.md](../RASPBERRY_PI.md) for detailed Raspberry Pi network configuration with dual interfaces.

#### 4. Base Path (`controlunit.base_path`)

**Default:** `"/var/opt/halko"`

Storage location for programs and execution logs. This directory is created by `make install`.

### Example Production Configurations

**Standard Linux server:**

```json
{
  "controlunit": {
    "base_path": "/var/opt/halko",
    "tick_length": "6s",
    "network_interface": "enp0s3",  // ← Change to your interface
    "defaults": { ... }
  },
  "power_unit": {
    "shelly_address": "http://192.168.1.50",  // ← Change to your Shelly IP
    "cycle_length": "60s",
    "max_idle_time": "70s",
    "power_mapping": { ... }
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",  // ← Verify your Arduino path
    "baud_rate": 9600
  },
  "api_endpoints": { ... }
}
```

**Raspberry Pi 3B (dual interface setup):**

```json
{
  "controlunit": {
    "base_path": "/var/opt/halko",
    "tick_length": "6s",
    "network_interface": "wlan0",  // WiFi for display IP
    "defaults": { ... }
  },
  "power_unit": {
    "shelly_address": "http://192.168.10.2",  // Shelly on eth0 static network
    "cycle_length": "60s",
    "max_idle_time": "70s",
    "power_mapping": { ... }
  },
  "sensorunit": {
    "serial_device": "/dev/ttyUSB0",  // or /dev/ttyACM0
    "baud_rate": 9600
  },
  "api_endpoints": { ... }
}
```

See [RASPBERRY_PI.md](../RASPBERRY_PI.md) for detailed Raspberry Pi deployment instructions.

## halko-daemon.service

Systemd service template for all Halko services. Used by `make systemd-units` to create:

- `halko@controlunit.service`
- `halko@powerunit.service`
- `halko@sensorunit.service`

The template uses systemd's instance unit pattern (`@`) to parameterize the service name.

## Post-Installation Steps

After running `make install` and `make systemd-units`:

1. **Edit configuration:**

   ```bash
   sudo nano /etc/opt/halko.cfg
   ```

   Update network_interface, serial_device, and shelly_address.

2. **Verify services started correctly:**

   ```bash
   sudo systemctl status halko@controlunit
   sudo systemctl status halko@powerunit
   sudo systemctl status halko@sensorunit
   ```

3. **Check logs for configuration errors:**

   ```bash
   sudo journalctl -u halko@controlunit -f
   ```

4. **Install webapp (optional):**

   ```bash
   make install-webapp
   sudo ln -s /etc/nginx/sites-available/halko /etc/nginx/sites-enabled/
   sudo systemctl reload nginx
   ```
