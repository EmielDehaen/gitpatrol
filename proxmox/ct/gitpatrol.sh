#!/usr/bin/env bash

# Copyright (c) 2021-2024 tteck (Maintained by community-scripts)
# License: MIT
# https://github.com/community-scripts/ProxmoxVE/raw/main/LICENSE

source /dev/stdin <<< "$({ wget -qLO - https://github.com/community-scripts/ProxmoxVE/raw/main/misc/api.func; } 2>/dev/null)"
source /dev/stdin <<< "$({ wget -qLO - https://github.com/community-scripts/ProxmoxVE/raw/main/misc/build.func; } 2>/dev/null)"

# Application Metadata
APP="GitPatrol"
var_disk="4"
var_cpu="1"
var_ram="512"
var_os="debian"
var_version="12"

# Function to handle updates (called via 'update' command in LXC)
update_script() {
  header_info
  if [[ ! -d /opt/gitpatrol ]]; then msg_error "No ${APP} Installation Found!"; exit; fi
  msg_info "Updating ${APP}"
  /opt/gitpatrol/scripts/install.sh --update
  msg_ok "Updated ${APP}"
  exit
}

# Header Info
header_info
echo -e "\nThis script will create a new ${APP} LXC.\n"

# Start the build process (Interactive)
build_container
description

# Final message
msg_ok "Completed Successfully!\n"
echo -e "${APP} is reachable at: http://${IP}:8080"
