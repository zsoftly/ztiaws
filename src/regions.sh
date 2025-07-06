#!/usr/bin/env bash

# AWS Regions mapping
# Format: "shortcode") echo "aws-region-name" ;;  # Location/City

get_region() {
    case "$1" in
        # Canada
        "cac1") echo "ca-central-1" ;;  # Montreal
        "caw1") echo "ca-west-1" ;;     # Calgary

        # United States
        "use1") echo "us-east-1" ;;     # N. Virginia
        "use2") echo "us-east-2" ;;     # Ohio
        "usw1") echo "us-west-1" ;;     # N. California
        "usw2") echo "us-west-2" ;;     # Oregon

        # Europe
        "euw1") echo "eu-west-1" ;;     # Ireland
        "euw2") echo "eu-west-2" ;;     # London
        "euw3") echo "eu-west-3" ;;     # Paris
        "euc1") echo "eu-central-1" ;;  # Frankfurt
        "eun1") echo "eu-north-1" ;;    # Stockholm
        "euc2") echo "eu-central-2" ;;  # Zurich
        "euw4") echo "eu-west-4" ;;     # Spain

        # Asia Pacific
        "aps1") echo "ap-south-1" ;;    # Mumbai
        "aps2") echo "ap-south-2" ;;    # Hyderabad
        "apse1") echo "ap-southeast-1" ;; # Singapore
        "apse2") echo "ap-southeast-2" ;; # Sydney
        "apse3") echo "ap-southeast-3" ;; # Jakarta
        "apse4") echo "ap-southeast-4" ;; # Melbourne
        "apne1") echo "ap-northeast-1" ;; # Tokyo
        "apne2") echo "ap-northeast-2" ;; # Seoul
        "apne3") echo "ap-northeast-3" ;; # Osaka
        "apne4") echo "ap-northeast-4" ;; # Hong Kong

        # South America
        "sae1") echo "sa-east-1" ;;     # São Paulo

        # Africa
        "afc1") echo "af-south-1" ;;    # Cape Town

        # Middle East
        "mec1") echo "me-central-1" ;;  # UAE
        "mec2") echo "me-central-2" ;;  # Saudi Arabia
        "mes1") echo "me-south-1" ;;    # Bahrain

        # Default
        *) echo "invalid" ;;
    esac
}

# Get region description
get_region_description() {
    case "$1" in
        "cac1") echo "Canada Central (Montreal)" ;;
        "caw1") echo "Canada West (Calgary)" ;;
        "use1") echo "US East (N. Virginia)" ;;
        "use2") echo "US East (Ohio)" ;;
        "usw1") echo "US West (N. California)" ;;
        "usw2") echo "US West (Oregon)" ;;
        "euw1") echo "EU West (Ireland)" ;;
        "euw2") echo "EU West (London)" ;;
        "euw3") echo "EU West (Paris)" ;;
        "euw4") echo "EU West (Spain)" ;;
        "euc1") echo "EU Central (Frankfurt)" ;;
        "euc2") echo "EU Central (Zurich)" ;;
        "eun1") echo "EU North (Stockholm)" ;;
        "aps1") echo "Asia Pacific South (Mumbai)" ;;
        "aps2") echo "Asia Pacific South (Hyderabad)" ;;
        "apse1") echo "Asia Pacific Southeast (Singapore)" ;;
        "apse2") echo "Asia Pacific Southeast (Sydney)" ;;
        "apse3") echo "Asia Pacific Southeast (Jakarta)" ;;
        "apse4") echo "Asia Pacific Southeast (Melbourne)" ;;
        "apne1") echo "Asia Pacific Northeast (Tokyo)" ;;
        "apne2") echo "Asia Pacific Northeast (Seoul)" ;;
        "apne3") echo "Asia Pacific Northeast (Osaka)" ;;
        "apne4") echo "Asia Pacific Northeast (Hong Kong)" ;;
        "sae1") echo "South America (São Paulo)" ;;
        "afc1") echo "Africa (Cape Town)" ;;
        "mec1") echo "Middle East (UAE)" ;;
        "mec2") echo "Middle East (Saudi Arabia)" ;;
        "mes1") echo "Middle East (Bahrain)" ;;
        *) echo "Unknown Region" ;;
    esac
}
