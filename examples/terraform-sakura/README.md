# Sakura Cloud Cost Estimation Example

This example covers the full range of Sakura Cloud compute resources supported by infracost.
This guide explains how to estimate costs and compare changes.

## File layout

Resources are split per use case so each section can be tweaked or removed independently:

| File | Resources |
|---|---|
| [`versions.tf`](./versions.tf) | terraform / provider configuration |
| [`network.tf`](./network.tf) | `sakura_internet`, `sakura_switch` |
| [`web.tf`](./web.tf) | Standard plan web servers + boot disks |
| [`db.tf`](./db.tf) | Dedicated Intel DB server + disks |
| [`batch.tf`](./batch.tf) | Dedicated AMD (EPYC) batch server |
| [`gpu.tf`](./gpu.tf) | GPU server (高火力VRT) |
| [`private_host.tf`](./private_host.tf) | Dedicated host (専有ホストプラン) |
| [`apprun.tf`](./apprun.tf) | AppRun shared (共有プラン) |
| [`ssh.tf`](./ssh.tf) | SSH key (free resource) |

## Architecture

| Resource | Spec | Monthly cost (excl. tax) |
|---|---|---|
| Router + Switch | 250Mbps | ¥24,000 |
| Switch (included) | — | ¥2,000 |
| Web server × 2 | Standard 4 core / 8GB | ¥8,400 × 2 |
| Web boot disk × 2 | SSD 100GiB | ¥3,500 × 2 |
| DB server | Dedicated Intel 8 core / 32GB | ¥46,400 |
| DB boot disk | SSD 40GiB | ¥1,400 |
| DB data disk | SSD 500GiB | ¥17,500 |
| Batch server | Dedicated AMD (EPYC 7003) 32 core / 120GB | ¥125,000 |
| GPU server | H100 (24 core / 240GB) | ¥350,000 |
| Dedicated host | dynamic | ¥200,000 |
| AppRun shared | 1 core / 2GiB × max 5 instances (upper bound) | ¥29,861 * |
| **Total** | | **≈ ¥818,561/month** |

`sakura_switch` and `sakura_ssh_key` are free resources and not included in the cost.  
`*` AppRun shared cost is a usage-based upper bound (see [AppRun shared](#apprun-shared-共有プラン) below).

## Prerequisites

- [infracost](https://www.infracost.io/docs/) is installed
- An infracost API key is configured (run `infracost auth login` to obtain one)

```sh
infracost auth login
```

## Running a Cost Breakdown

Sakura Cloud pricing is denominated in **JPY**, so pass `INFRACOST_CURRENCY=JPY` when running infracost.

```sh
INFRACOST_CURRENCY=JPY infracost breakdown --path .
```

### Example output

```
Project: main

 Name                                                          Monthly Qty  Unit    Monthly Cost (JPY)

 sakura_server.gpu
 └─ Server GPU H100 (×1)                                               1  months            ¥350,000

 sakura_private_host.host
 └─ Private host (dynamic)                                             1  months            ¥200,000

 sakura_server.batch
 └─ Server dedicated AMD (32 core, 120GB)                              1  months            ¥125,000

 sakura_server.db
 └─ Server dedicated Intel (8 core, 32GB)                              1  months             ¥46,400

 sakura_apprun_shared.api
 └─ AppRun shared (1.0 core, 2.0GiB × max 5 instances, upper bound)   1  months             ¥29,861 *

 sakura_internet.main
 ├─ ルータ+スイッチ (250Mbps)                                          1  months             ¥24,000
 └─ スイッチ                                                           1  months              ¥2,000

 sakura_disk.db_data
 └─ Disk (ssd, 500GiB)                                                 1  months             ¥17,500

 sakura_server.web[0]
 └─ Server (4 core, 8GB)                                               1  months              ¥8,400

 sakura_server.web[1]
 └─ Server (4 core, 8GB)                                               1  months              ¥8,400

 sakura_disk.web_boot[0]
 └─ Disk (ssd, 100GiB)                                                 1  months              ¥3,500

 sakura_disk.web_boot[1]
 └─ Disk (ssd, 100GiB)                                                 1  months              ¥3,500

 OVERALL TOTAL (JPY)                                                                        ¥818,561
```

## Comparing Costs After a Change

Edit any `.tf` file (lines marked with `<<<`) and run `infracost diff` to see the cost impact:

```sh
INFRACOST_CURRENCY=JPY infracost diff --path .
```

## Supported Resources

### `sakura_server` — Server

Plan is determined by the combination of `core`, `memory`, `commitment`, `cpu_model`, and `gpu`.

**Standard plan** (通常プラン)

```hcl
resource "sakura_server" "example" {
  core   = 4
  memory = 8
}
```

**Dedicated core — Intel** (コア専有プラン / Intel Xeon)

```hcl
resource "sakura_server" "example" {
  core       = 8    # 2 / 4 / 6 / 8 / 10 / 24
  memory     = 32   # varies by core count
  commitment = "dedicated"
}
```

**Dedicated core — AMD** (コア専有プラン / AMD EPYC)

```hcl
resource "sakura_server" "example" {
  core       = 32        # 32 / 64 / 128 / 192
  memory     = 120       # 120 / 240 / 480 / 1024
  commitment = "dedicated"
  cpu_model  = "epyc_7003"   # epyc_7003 / epyc_9004
}
```

**GPU server — 高火力VRT** (Ishikari Zone 1 only)

```hcl
resource "sakura_server" "example" {
  gpu       = 1
  gpu_model = "H100"   # H100 / V100
}
```

---

### `sakura_disk` — Disk

```hcl
resource "sakura_disk" "example" {
  plan = "ssd"   # ssd / hdd
  size = 100     # GiB
}
```

---

### `sakura_internet` — Router + Switch

```hcl
resource "sakura_internet" "example" {
  band_width = 250   # Mbps: 100 / 250 / 500 / 1000 / ...
}
```

---

### `sakura_private_host` — Dedicated Host (専有ホストプラン)

```hcl
resource "sakura_private_host" "example" {
  class = "dynamic"   # dynamic / windows
}
```

| class | Monthly cost (excl. tax) |
|---|---|
| `dynamic` | ¥200,000 |
| `windows` | ¥230,000 |

---

### `sakura_apprun_shared` — AppRun Shared (共有プラン)

Cost is a **usage-based upper bound**: `hourly_rate × 730h × max_scale`.  
Actual cost is lower when instances scale to zero during low-traffic periods.

```hcl
resource "sakura_apprun_shared" "example" {
  min_scale = 0   # 0 enables scale-to-zero
  max_scale = 5

  components = [{
    max_cpu    = "1"    # 0.5 / 1 / 2
    max_memory = "2Gi" # 1Gi / 2Gi / 4Gi
    ...
  }]
}
```

| max_cpu | max_memory | Hourly rate (excl. tax) |
|---|---|---|
| 0.5 | 1Gi | ¥4.5 |
| 1 | 1Gi | ¥7.3 |
| 1 | 2Gi | ¥8.2 |
| 2 | 2Gi | ¥15.5 |
| 2 | 4Gi | ¥17.3 |

---

### Free resources (no cost)

`sakura_switch`, `sakura_ssh_key`, `sakura_packet_filter`, `sakura_script`,
`sakura_icon`, `sakura_bridge`, `sakura_auto_backup`, `sakura_vswitch`,
`sakura_ipv4_ptr`, `sakura_apprun_dedicated_cluster`

## Pricing Notes

- Prices are based on the [Sakura Cloud official pricing page](https://cloud.sakura.ad.jp/products/) and are **tax-excluded**
- Prices reflect the Ishikari Zone 1 (`is1a`) rate
- Actual charges may vary depending on the zone and any discount passport applied
