# jorvit

Tools to manage and test self VIT node operations and integrations.

## Installation

1. Download the latest release from releases page [https://github.com/input-output-hk/jorvit/releases](https://github.com/input-output-hk/jorvit/releases)
2. Download the jörmungandr nightly binaries from [https://github.com/input-output-hk/jormungandr/releases/tag/nightly](https://github.com/input-output-hk/jormungandr/releases/tag/nightly) and place them inside `jor_bins` folder

### Usage

Upon installation the content/structure of the working folder should be like:

```sh
.
├── assets
│   ├── fund.csv       <-- fund general info
│   └── proposals.csv  <-- proposals list example
├── jor_bins
│   ├── jcli         <-- node binary
│   ├── jormungandr  <-- node binary
│   └── Readme.md
├── jorvit           <-- our binary
└── README.md

2 directories, 7 files
```

Just execute `jorvit` binary within a shell and the application will start in foreground
and will display some informations related to the content and services it provides.

```log
➜  jorvit_Linux_x86_64 ./jorvit
2020/05/26 19:38:35 Proposals File load took 708.645µs
2020/05/26 19:38:35 Fund File load took 120.893µs
2020/05/26 19:38:35 Working Directory: /tmp/jorvit_Linux_x86_64/jnode_VIT_741709330
2020/05/26 19:38:35
2020/05/26 19:38:35 OS: linux, ARCH: amd64
2020/05/26 19:38:35
2020/05/26 19:38:35 jcli: /tmp/jorvit_Linux_x86_64/jor_bins/jcli
2020/05/26 19:38:35 ver : jcli 0.9.0-nightly (master-c87061eb, release, linux [x86_64]) - [rustc 1.43.1 (8d69840ab 2020-05-04)]
2020/05/26 19:38:35
2020/05/26 19:38:35 node: /tmp/jorvit_Linux_x86_64/jor_bins/jormungandr
2020/05/26 19:38:35 ver : jormungandr 0.9.0-nightly (master-c87061eb, release, linux [x86_64]) - [rustc 1.43.1 (8d69840ab 2020-05-04)]
2020/05/26 19:38:35
2020/05/26 19:38:35 VIT - BFT Genesis Hash: 9c1b9b82e86faeb43dceaa2008a2c5dded07b2b66ed0469a9fd213c262242534
2020/05/26 19:38:35
2020/05/26 19:38:35 VIT - BFT Genesis: COMMITTEE - 2
2020/05/26 19:38:35 VIT - BFT Genesis: VOTEPLANS - 2
2020/05/26 19:38:35 VIT - BFT Genesis: PROPOSALS - 20
2020/05/26 19:38:35
2020/05/26 19:38:35 VIT - BFT Genesis: Wallets available for recovery


W: DAEDALUS
M: tired owner misery large dream glad upset welcome shuffle eagle pulp time
A: 20101
█████████████████████████████████████████████
█████████████████████████████████████████████
████ ▄▄▄▄▄ ██▀█  ▀████ ▀▄ ▄▄▀█▀▀▄█ ▄▄▄▄▄ ████
████ █   █ █ ▄███ ▄▄▄▀█ ▄▄▀▀▀▀  ██ █   █ ████
████ █▄▄▄█ █ ██▄█▀█▀██ ▄▄ █▄▄█ ▀ █ █▄▄▄█ ████
████▄▄▄▄▄▄▄█ ▀ █ ▀▄█▄█▄█ ▀▄█▄█ █▄█▄▄▄▄▄▄▄████
████▄▀   ▄▄██▄▀▀█▄▄▀▀█▄▄▄▀█▀▄▄▄▄██▄ ▄▄ ▀█████
█████ ▄▄█▄▄█▀▄▄ ▀▀▄█▀█▀█▄ ▄██▄█  ▄█▄▄▄█  ████
████▀▄█▀▀█▄▀ ▀▀▄ ▄▄██  ▀▀▀ ▀▄▄ ▄ ▄▄▀▄ ▄██████
████▄▀▀ █ ▄▄▀█▄▄▄█  ▀ ▀▀  █▀██▀ ▄█▀ ▄▄▄  ████
████▄▀ ▀█ ▄▀██▀▀▀ █ ▄█▀█▀▀ ▀█ ▄█▀ ▄ ▄ ▄▄▄████
████▄▀  ▄▀▄▄ █ ▀▄▄█ ▄█  ▀▀██ █▀ ▀  ███▄▄ ████
████▄ ▀▄▄▄▄▀▀▄▀██▄ ▄ ▄█ ▀▀▄▀▄▄██▀ ▄ ▄▄▄▄▄████
████▀██▄ ▄▄  █▀▀▄▄█  █▄██ ▄▀ ▄       ██ ▄████
████ ▀▄▀  ▄ ▄ █▄    ▀█▄██▀▄▀█▄▄█▀▄▄█▄▄  ▄████
████ █▀▀█ ▄▀ ▄▄ █ ▀█▀███   █ ██▄▀█▄ ██▄  ████
████▄███▄█▄▄ ▄▀▄ ▀▄█▀ █▀██▄▀█ ▄  ▄▄▄ ▄▄▀█████
████ ▄▄▄▄▄ █▀▀██ ▀█▄▄ ▀█  ▀█▀█▄█ █▄█  █▄ ████
████ █   █ █ █▄▄█▄█ ▄█▀ ▀▀▄▀  ▄ ▄▄ ▄ ▄ █ ████
████ █▄▄▄█ █▄▄   █▄ ▄█ █  ▀███▄ ▀▀▀▄███▄ ████
████▄▄▄▄▄▄▄█▄▄█▄▄▄▄▄▄▄█▄██▄█▄▄▄██████▄▄▄▄████
█████████████████████████████████████████████
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀



W: DAEDALUS
M: edge club wrap where juice nephew whip entry cover bullet cause jeans
A: 101001
█████████████████████████████████████████████
█████████████████████████████████████████████
████ ▄▄▄▄▄ █▄█▄▄▀▄ █▀█ ▄▀▀█▄█▄▄▄▄█ ▄▄▄▄▄ ████
████ █   █ █▀█ ▀█▄▀ █▄▀ ▄▄█ █▀  ▄█ █   █ ████
████ █▄▄▄█ █ █ █▄▄▄▀██▀▀▄ ▀▄█▄ ▄▄█ █▄▄▄█ ████
████▄▄▄▄▄▄▄█ █▄▀ █▄█▄▀ ▀ ▀ ▀▄█ ▀ █▄▄▄▄▄▄▄████
████ ██▀▄▀▄ ▄█ █▄▀ █▀█▄ ▀▄  ▄█▀▀█▄▄▄ ▄██▄████
████▀▄█▀▄ ▄▄▄█▄ █ ▀█▄▀ █▄ █  ██ ▄ ▀██▄ ▀█████
████▄█▀ ██▄   █▀ ▄█   ▀▄█▄█▀▄▀▀▀██▄▄▀▀▄▄▄████
█████▄█ ▄▄▄▀██▄▄▄▄█▄▄█▄█ ▀▄  █▀   ▀ ██▀▄█████
████▀▄▄▄▀█▄▀██▀  █ ▄▄▀▀ ▄██▀▄▀█▀  ▄ ▀█▄██████
████ █▀█ ▄▄▄▄▀ █▀▄▀ ▀ █ ▀  ▄██  █ ██▄█▀██████
████▀ █▄█ ▄▀▀▄▀▀▄ █▀▀▄▀█▄ ▄ █▀█▀  ▄█▀▀▄ ▄████
████▄█▀█  ▄ █▀ ▄▀▀▄ █ ▀███▀▄██  ▀██▀▀▀ ▀▀████
████▀▄▀█▀█▄▄ ▀▀▀▀▄ █▀▀▄▀▀▀▄ ▄▀█▀▀▄▄█▀▀ █▄████
██████▄▄▄█▄ ████▀ ▀█▄ ▀█▄▀▀ █▄█ ▄█▀ ██▀██████
████▄▄▄███▄█ ▄▀▀▄ █ ▀ █▄▄▄▄▀▄▀▀▄ ▄▄▄ ▄▄ ▄████
████ ▄▄▄▄▄ █▄██▄▄▄▀  █▄██ ▄ █▄▀█ █▄█ ▄ ▀█████
████ █   █ █▄▀██▄▀▄█▄█▀ ▄█▀▀ ▀▀█    ▄▀▄▄ ████
████ █▄▄▄█ ██  █ ██ ▀ █ ▀ █  █  ▄█  ▀▄▀▀█████
████▄▄▄▄▄▄▄█▄██▄▄████▄▄██▄▄█▄█▄▄██▄▄▄█▄▄▄████
█████████████████████████████████████████████
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀



W: YOROI
M: neck bulb teach illegal soul cry monitor claw amount boring provide village rival draft stone
A: 30001
█████████████████████████████████████████████████
█████████████████████████████████████████████████
████ ▄▄▄▄▄ █  █▄▄▄█▀▄▄ ▄ ▄█ █ ██▀█▄▄ █ ▄▄▄▄▄ ████
████ █   █ █▀▄▀█ ▀ ██▄███▀█▀ █▀▀█▀▄███ █   █ ████
████ █▄▄▄█ ██▄  ▀ ▀ ▀▄█ ▀▄▄█▀█▄▄▀ ▄ ▀█ █▄▄▄█ ████
████▄▄▄▄▄▄▄█ ▀ █ █▄▀▄▀▄█▄█▄█ █▄▀▄▀▄▀▄█▄▄▄▄▄▄▄████
████▄▀  ▀ ▄ ███ █ ███ ▄ █ █▀██▄█  ▀▀▀▀ ▀▀ ▀  ████
██████▀▄ █▄█▀█▀▀█▀████▄▀ █▀▄▀▀▄█ █▄▀ ▀▀  ███▀████
█████▄█ ▀█▄ ▀ ▄ █▄▀▄ ▄▀▄ █▄  ▄ ▄█▄█ ▄█▀▄█▀▄█ ████
████ ▀█ ██▄▀▀▀█▄▀ ▀█▀██▄▄▀  █▄█▀▀▄▄▄█▀▄█ ▄█  ████
█████▄▄  ▀▄█▄ ▀▀▄  █▄█ █▀  ▀▀ ▀▀██ █▀▄█▄▄█▀██████
████▄ ▄█▀▄▄█▄▄ ██ █  ▀ █▀▀▄▀ ██▄▀▀██▀ ▄ ▀  ▀█████
█████  █  ▄▄█▄▄  ▄█ ▄██  ▀█ ▄ ▀█▀▄  ▀▀ ▀ █▄  ████
█████▀▄ █▄▄█▄▄ ▄▀▄▄█▄█▀█▀█▄▄▄▀▄▀▄▄█▀▀ █▄▄█▄ ▀████
████  ▄▄▀▀▄▄▀▄██▄▀▀▀  █ ██▄█ ▄▀▀▀▄▀ █▄▀██▀ █ ████
████▀▀ ▀▀█▄▄▄▀ ▀█ ▀█▀█ ▄▄▀ █ █▀▀▀██▄   ▄█▀█▄▄████
█████▄█▀█ ▄█  ▀▀▄ █▄▀██▀ ▄ █▄ █▀██ █▄ █▀▄▀███████
████▄█▄▀▄▀▄▀▄▀▀▄█▀▄▀█ █ ██ ▀█▄  ▀ ███▄▀ █▀▀██████
█████▄███▄▄▄▀ █▄▄ ▄ ▄  ▀█ █▀█▀▄▄     ▄▄▄ ▄▄█ ████
████ ▄▄▄▄▄ █  ▄▄ ▄ █▀▀▄▀▀██▄▄▄██ █ ▄ █▄█ ██ █████
████ █   █ █▀█▀▄█▀█████▀█▄▀▄ █ ▀▀ ██  ▄  ▀ ▀▄████
████ █▄▄▄█ █▄▄ ▀▀▀██▄▄▀▄  ▀██▄█▀ █ ▄▀ █ █ █ ▄████
████▄▄▄▄▄▄▄█▄▄▄▄██▄▄▄▄▄▄▄▄▄██▄████▄▄▄██▄█▄█▄█████
█████████████████████████████████████████████████
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀



W: DAEDALUS-PAPER
M: town lift more follow chronic lunch weird uniform earth census proof cave gap fancy topic year leader phrase state circle cloth reward dish survey act punch bounce
A: 40000
█████████████████████████████████████████████████████████████
█████████████████████████████████████████████████████████████
████ ▄▄▄▄▄ ██ ▄ ▄▄██ ▄  ▄█▀▀ ▀█▄▄ ▄▄▄▀▀▀▄▄▀ ▄▀ ▀██ ▄▄▄▄▄ ████
████ █   █ █  ▀    ▄██▀ █▄▄▀▄▀█▄█▄▀ ▄ ▄█  █ ▄▀▀▄▀█ █   █ ████
████ █▄▄▄█ █ █▀▀ █▀▄▀   █ ▄▄ ▄▄▄ ▄▄█▀ ▀  █▀██▄ ███ █▄▄▄█ ████
████▄▄▄▄▄▄▄█ ▀ ▀ ▀ █ █ ▀▄▀▄█ █▄█ █ █▄▀ █▄█▄▀ ▀ ▀▄█▄▄▄▄▄▄▄████
████ █ ▄  ▄██▀█ █▀█▀▀ ▄▄█▀▀█ ▄▄▄   ▄ ██▄ ▀█▀██ ▀▄█▄▄  ▄▀▀████
█████ ▀ ▀█▄▀▄  █▄█ ▄█ ▀▄▄ ▀ ▀▄▄██▄▀▄    ▄▀ ▀▀██▄▄▄▄▀▄ ▄▀█████
████ ▄▀█▄█▄ ██▀█▄ █▀▄█ ▄█▀▀█▄█▄█▀▄  ▀ █▄  ▄▄█  ▀▄ █ ▄▀▄█▀████
████▄▄▄█ ▀▄▄ █ ▀▀ ▄▀▀▀▀▄▀▄▀█▀▀▄▀█▄ ▄▀▄ █▄▄▄▀ ▀█▄█ ▄▀▄▄███████
████▄███▀▄▄ ▀ ▄▄█▄▀▀▀▄█▀ █▀▀▀█▄ ▀▀▀▄█▄ ▄ ▀█▀▀▀▀▀▄ █  █▄▀▀████
████ █▀▀  ▄▄▀ ▄▄▄ █▀▄▄▀▀█▀ ▀▀█▄██▀▀▀▀▄▀▄▄▀  ▀██▀ ▄▄▀▄▀█▄█████
████▀▄█▀█▀▄▄▀█▀█  █▄   ▀▀▀▄  █  ▄  ▄▀█▄▄  ▄▀█   ▄▄█  ▄ ▀▀████
████▄█  █ ▄▀ █▀▀ ▀██▄███▄▄▀▀▄ █▀  ▀▄ ▀▄▄█▄▄ ▀ █▀▄▄▄█▄▄█▀▀████
█████▀█▀ ▄▄▄ █ ▄ ▄▄▀█▄▄█ ▀▀▀ ▄▄▄    ▀▄█ ▄▀█▀█▀▀  ▄▄▄ ▀ █▀████
████ █ █ █▄█  █▀█▄ ▄█▀▄█ ▄▀  █▄█ █▀▀█▄▄▀▄▀▄█ █▀▄ █▄█  ▄▀█████
████▀  ▄▄▄▄▄▄ █  █ █▀▀█▄ ▀▀▀   ▄ ▀▀▀███ ▄ █▀█▀ █▄ ▄   ▄▄▀████
████  █▄  ▄▄ ▀█▄  ▄▄▀▀▀▄ ▄██ ▄ ▄▄▀█▀▀ ▄▄██▄ ▀▀▀▀▀ ▀█▀ ██▀████
████▄▀▄ ▀█▄█ █▄▄ ▄█▀▄█▀███▀ ▄▄▄ █▀█ █▄█▀▄▀█▀▀▀▀██▀▄█▀█▄ ▀████
████ ▄▄▄▄▄▄   █▀   ▄▄ ▄▄▀███ ██▄▄▀▄▄▀▀▄▀▄█▄▄▀█▀▀▀▀▀▀█▀█▀▀████
████▄█▀  ▄▄▄█  █▄▀█▀█  ███▀▀▄▄▀▄█ ▀█▀▀█▄▄ ▄█▀ ▀▀▀ █▀▀█▄▀▀████
████▄▀ ▄▄ ▄  █ █▄▀▀ ▄  ▀ ▄█▀█  █▄█▀█▀███▄▀▄▀ ▄▀▀▀█▀▄  ▄█▀████
████▀▄▀▄▀▀▄▄ █▄▄▄ ▄▄ ▀▀▀▀▄█ █▄▀ █▄▀ █▀▄ ▄ █ █▀▀█  █▀▀▄▄▄▀████
████▄ ▀▄▄▄▄ ▀█ █ █▄▀▀█▀▄█ ▀█▄██▄▄█▀▄▀ ▄▄█▀▄  █▀▀▀▀▀▀ ▀███████
███████▄██▄█▀██ █▀▄█▀ ▄ ▀█▀  ▄▄▄   ▄██▄▄▄▀▄█▀ ▀▀ ▄▄▄ ▀▄ ▀████
████ ▄▄▄▄▄ █▀▀ ▄▄ ▀█▀▀ ▄▀▀ █ █▄█ ▀▀██ ▄▄██▄ ▀▀██ █▄█  ▄ █████
████ █   █ █ ▄▀▄▀▀██▄▄▄ █▀▄▀  ▄  █ ▄ ██▄  █▄█    ▄▄▄▄█   ████
████ █▄▄▄█ █▄▄█▀ █▀█▄▀██▀ ██▄▄▄█▄█▀▄ ▀▄▄█▄ ▄ ▀█▄█▀█ ▄ █ ▄████
████▄▄▄▄▄▄▄█▄▄█▄▄██████▄███▄▄▄▄▄█▄▄▄█▄▄▄▄▄███▄███▄▄▄▄▄▄██████
█████████████████████████████████████████████████████████████
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀

2020/05/26 19:38:35
2020/05/26 19:38:35 JÖRMUNGANDR listening at: /ip4/127.0.0.11/tcp/9001
2020/05/26 19:38:35 JÖRMUNGANDR Rest API available at: http://0.0.0.0:8001/api
2020/05/26 19:38:35
2020/05/26 19:38:35 APP - PROXY Rest API available at: http://0.0.0.0:8000/api
2020/05/26 19:38:35
2020/05/26 19:38:35 VIT - BFT Genesis Node - Running...

```

### APP - PROXY Rest API

The important service is the `APP - PROXY Rest API` since the other 2 services are provided from the jörmungandr service itself.

This service exposes the following rest endpoints:

1. `/api/v0/fund` - provides general info related to the active fund phase:

   ```sh
   curl 'http://localhost:8000/api/v0/fund'
   ```

   ```json
   {
     "fund_name": "Fund0",
     "fund_goal": "Support innovative projects",
     "voting_power_info": "2020-05-26 15:26:29 +0000 UTC",
     "rewards_info": "2020-06-26 15:26:29 +0000 UTC",
     "fund_start_time": "2020-05-26 15:26:29 +0000 UTC",
     "fund_end_time": "2020-06-26 15:26:29 +0000 UTC",
     "next_fund_start_time": "2021-06-26 15:26:29 +0000 UTC",
     "chain_vote_plans": [
       {
         "chain_voteplan_id": "f4fdab54e2d516ce1cabe8ae8cfe77e99eeb530f7033cdf20e2392e012373a7b",
         "chain_vote_starttime": "2020-05-26 19:38:35 +0000 UTC",
         "chain_vote_endtime": "2020-05-26 19:43:35 +0000 UTC",
         "chain_committee_endtime": "2020-05-26 19:48:35 +0000 UTC",
         "chain_voteplan_payload": "Public"
       },
       {
         "chain_voteplan_id": "145b208b9de264352ae8b7071ee1c59996de01dde03ce8ecd5b44f7f71631cec",
         "chain_vote_starttime": "2020-05-26 19:38:35 +0000 UTC",
         "chain_vote_endtime": "2020-05-26 19:43:35 +0000 UTC",
         "chain_committee_endtime": "2020-05-26 19:48:35 +0000 UTC",
         "chain_voteplan_payload": "Public"
       }
     ]
   }
   ```

2. `/api/v0/proposals/{internal_id}` - get a single proposal details based on `internal_id`:

   ```sh
   curl 'http://localhost:8000/api/v0/proposals/1'
   ```

   ```json
   {
     "internal_id": "1",
     "category": {
       "category_id": "",
       "category_name": "Fund0 Development",
       "category_description": ""
     },
     "proposal_id": "16444246",
     "proposal_title": "Test proposal 16444246",
     "proposal_summary": "To test the proposal process 16444246",
     "proposal_problem": "We haven't tested proposal integration yet 16444246",
     "proposal_solution": "Test the proposal integration process 16444246",
     "proposal_url": "https://iohk.submittable.com/submissions/16444246",
     "proposal_files_url": "https://iohk.submittable.com/submissions/16444246/file/0",
     "proposal_public_key": "Ae2tdPwUPEYwrazXRJVK4NgHSZCjP9kLSMrx2awgYiBH61zT8kz6u33Sije",
     "proposal_funds": 1000246,
     "proposer": {
       "proposer_email": "iohk_16444246@iohk.io",
       "proposer_name": "IOHK 16444246",
       "proposer_url": "https://iohk.io"
     },
     "chain_proposal_id": "5db05d3c7bfc37f2059d24966aa6ef05cfa25b6a478dedb3b93f5dca5c57c24a",
     "chain_proposal_index": 0,
     "chain_vote_options": {
       "blank": 0,
       "YES": 1,
       "NO": 2
     },
     "chain_voteplan_id": "f4fdab54e2d516ce1cabe8ae8cfe77e99eeb530f7033cdf20e2392e012373a7b",
     "chain_vote_starttime": "2020-05-26 19:38:35 +0000 UTC",
     "chain_vote_endtime": "2020-05-26 19:43:35 +0000 UTC",
     "chain_committee_endtime": "2020-05-26 19:48:35 +0000 UTC",
     "chain_voteplan_payload": "Public"
   }
   ```

3. `/api/v0/proposals` - get a array with all detailed proposals:

   ```sh
   curl 'http://localhost:8000/api/v0/proposals'
   ```

   ```json
   [
     {
       "internal_id": "1",
       "category": {
         "category_id": "",
         "category_name": "Fund0 Development",
         "category_description": ""
       },
       "proposal_id": "16444246",
       "proposal_title": "Test proposal 16444246",
       "proposal_summary": "To test the proposal process 16444246",
       "proposal_problem": "We haven't tested proposal integration yet 16444246",
       "proposal_solution": "Test the proposal integration process 16444246",
       "proposal_url": "https://iohk.submittable.com/submissions/16444246",
       "proposal_files_url": "https://iohk.submittable.com/submissions/16444246/file/0",
       "proposal_public_key": "Ae2tdPwUPEYwrazXRJVK4NgHSZCjP9kLSMrx2awgYiBH61zT8kz6u33Sije",
       "proposal_funds": 1000246,
       "proposer": {
         "proposer_email": "iohk_16444246@iohk.io",
         "proposer_name": "IOHK 16444246",
         "proposer_url": "https://iohk.io"
       },
       "chain_proposal_id": "5db05d3c7bfc37f2059d24966aa6ef05cfa25b6a478dedb3b93f5dca5c57c24a",
       "chain_proposal_index": 0,
       "chain_vote_options": {
         "blank": 0,
         "YES": 1,
         "NO": 2
       },
       "chain_voteplan_id": "f4fdab54e2d516ce1cabe8ae8cfe77e99eeb530f7033cdf20e2392e012373a7b",
       "chain_vote_starttime": "2020-05-26 19:38:35 +0000 UTC",
       "chain_vote_endtime": "2020-05-26 19:43:35 +0000 UTC",
       "chain_committee_endtime": "2020-05-26 19:48:35 +0000 UTC",
       "chain_voteplan_payload": "Public"
     },
     {
       "internal_id": "20",
       "category": {
         "category_id": "",
         "category_name": "Fund0 Development",
         "category_description": ""
       },
       "proposal_id": "16444265",
       "proposal_title": "Test proposal 16444265",
       "proposal_summary": "To test the proposal process 16444265",
       "proposal_problem": "We haven't tested proposal integration yet 16444265",
       "proposal_solution": "Test the proposal integration process 16444265",
       "proposal_url": "https://iohk.submittable.com/submissions/16444265",
       "proposal_files_url": "https://iohk.submittable.com/submissions/16444265/file/0",
       "proposal_public_key": "Ae2tdPwUPEYwrazXRJVK4NgHSZCjP9kLSMrx2awgYiBH61zT8kz6u33Sije",
       "proposal_funds": 1000265,
       "proposer": {
         "proposer_email": "iohk_16444265@iohk.io",
         "proposer_name": "IOHK 16444265",
         "proposer_url": "https://iohk.io"
       },
       "chain_proposal_id": "31a4ecb01eeae808323a11621173f684f64cd35b76b5fe876abfaf694095fee9",
       "chain_proposal_index": 9,
       "chain_vote_options": {
         "blank": 0,
         "YES": 1,
         "NO": 2
       },
       "chain_voteplan_id": "145b208b9de264352ae8b7071ee1c59996de01dde03ce8ecd5b44f7f71631cec",
       "chain_vote_starttime": "2020-05-26 19:38:35 +0000 UTC",
       "chain_vote_endtime": "2020-05-26 19:43:35 +0000 UTC",
       "chain_committee_endtime": "2020-05-26 19:48:35 +0000 UTC",
       "chain_voteplan_payload": "Public"
     }
   ]
   ```

4. `/api/v0/block0` - get the binary content of the genesis block needed for wallet recovery:

   ```sh
   curl 'http://localhost:8000/api/v0/block0'
   ```

#### Additionals

There are also some endpoints **proxied** to the Jörmungadr node Rest service.
Are provided in the proxy endpoint for convenience, just to keep the same endpoint in the client.
The api docs can be found [here](https://editor.swagger.io/?url=https://raw.githubusercontent.com/input-output-hk/jormungandr/master/doc/api/v0.yaml)

- `/api/v0/account`  - needed ti update the wallet state
- `/api/v0/message`  - needed to send transactions (wallet recovery, voting, ...)
- `/api/v0/fragment` -
- `/api/v0/settings` -
- `/api/v0/block`    -
