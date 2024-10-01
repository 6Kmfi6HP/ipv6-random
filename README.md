# ipv6-random

This repository contains scripts for setting up and configuring network services with random IPv6 support.

## Quick Start

To quickly set up the environment, you can use one of the following commands:

```bash
wget -O install.sh https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/install.sh && chmod +x install.sh && ./install.sh
```

If you prefer not to save the downloaded file, you can use this command instead:

```bash
bash <(curl -s https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/install.sh)
```

## Usage

### 1. Install Dependencies

Before running the script, ensure that the following dependencies are installed:

- `iptables`
- `ip6tables`
- `netfilter-persistent`

### 2. Run the Script

To run the script, use the following command:

```bash
./install.sh
```

### 3. Configure Services

After running the script, you will be prompted to configure the services you want to set up. You can choose from the following options:

- **Vmess**: A secure and efficient proxy protocol.
- **Socks**: A proxy protocol for secure and efficient connections.

For each service, you will be asked to enter the port range and other relevant details.

### 4. Start Services

To start the services, use the following command:

```bash
systemctl start xrayK.service
```

### 5. Verify Configuration

To verify the configuration, you can check the status of the services using the following command:

```bash
systemctl status xrayK.service
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
