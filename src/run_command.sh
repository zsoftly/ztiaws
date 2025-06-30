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
        "euc2") echo "eu-central-2" ;;  # Zurich

        # Asia Pacific
        "aps1") echo "ap-south-1" ;;    # Mumbai
        "aps2") echo "ap-southeast-3" ;; # Jakarta
        "apse1") echo "ap-southeast-1" ;; # Singapore
        "apse2") echo "ap-southeast-2" ;; # Sydney
        "apse4") echo "ap-southeast-4" ;; # Melbourne
        "apne1") echo "ap-northeast-1" ;; # Tokyo
        "apne2") echo "ap-northeast-2" ;; # Seoul
        "apne3") echo "ap-northeast-3" ;; # Osaka

        # South America
        "sae1") echo "sa-east-1" ;;     # São Paulo

        # Africa
        "afs1") echo "af-south-1" ;;    # Cape Town

        # Middle East
        "mec1") echo "me-central-1" ;;  # Bahrain
        "mee1") echo "me-east-1" ;;     # UAE
        "meil1") echo "me-il-1" ;;      # Israel (Tel Aviv)

        # Default
        *) echo "invalid" ;;
    esac
}

# Get region description
get_region_description() {
    case "$1" in
        # Canada
        "cac1") echo "Canada Central (Montreal)" ;;
        "caw1") echo "Canada West (Calgary)" ;;

        # United States
        "use1") echo "US East (N. Virginia)" ;;
        "use2") echo "US East (Ohio)" ;;
        "usw1") echo "US West (N. California)" ;;
        "usw2") echo "US West (Oregon)" ;;

        # Europe
        "euw1") echo "Europe (Ireland)" ;;
        "euw2") echo "Europe (London)" ;;
        "euw3") echo "Europe (Paris)" ;;
        "euc1") echo "Europe (Frankfurt)" ;;
        "euc2") echo "Europe (Zurich)" ;;

        # Asia Pacific
        "aps1") echo "Asia Pacific (Mumbai)" ;;
        "aps2") echo "Asia Pacific (Jakarta)" ;;
        "apse1") echo "Asia Pacific (Singapore)" ;;
        "apse2") echo "Asia Pacific (Sydney)" ;;
        "apse4") echo "Asia Pacific (Melbourne)" ;;
        "apne1") echo "Asia Pacific (Tokyo)" ;;
        "apne2") echo "Asia Pacific (Seoul)" ;;
        "apne3") echo "Asia Pacific (Osaka)" ;;

        # South America
        "sae1") echo "South America (São Paulo)" ;;

        # Africa
        "afs1") echo "Africa (Cape Town)" ;;

        # Middle East
        "mec1") echo "Middle East (Bahrain)" ;;
        "mee1") echo "Middle East (UAE)" ;;
        "meil1") echo "Middle East (Tel Aviv)" ;;

        # Default
        *) echo "Unknown Region" ;;
    esac
}
