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

        # South America
        "sae1") echo "sa-east-1" ;;     # São Paulo

        # Europe
        "euw1") echo "eu-west-1" ;;     # Ireland
        "euw2") echo "eu-west-2" ;;     # London
        "euw3") echo "eu-west-3" ;;     # Paris
        "euc1") echo "eu-central-1" ;;  # Frankfurt
        "euc2") echo "eu-central-2" ;;  # Zurich
        "eun1") echo "eu-north-1" ;;    # Stockholm
        "eus1") echo "eu-south-1" ;;    # Milan
        "eus2") echo "eu-south-2" ;;    # Spain

        # Asia Pacific
        "aps1") echo "ap-south-1" ;;      # Mumbai
        "aps2") echo "ap-south-2" ;;      # Hyderabad
        "apse1") echo "ap-southeast-1" ;; # Singapore
        "apse2") echo "ap-southeast-2" ;; # Sydney
        "apse3") echo "ap-southeast-3" ;; # Jakarta
        "apse4") echo "ap-southeast-4" ;; # Melbourne
        "apne1") echo "ap-northeast-1" ;; # Tokyo
        "apne2") echo "ap-northeast-2" ;; # Seoul
        "apne3") echo "ap-northeast-3" ;; # Osaka

        # Middle East
        "mec1") echo "me-central-1" ;;    # UAE
        "mees1") echo "me-south-1" ;;     # Bahrain

        # Africa
        "afc1") echo "af-south-1" ;;      # Cape Town

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

        # South America
        "sae1") echo "South America (São Paulo)" ;;

        # Europe
        "euw1") echo "Europe (Ireland)" ;;
        "euw2") echo "Europe (London)" ;;
        "euw3") echo "Europe (Paris)" ;;
        "euc1") echo "Europe Central (Frankfurt)" ;;
        "euc2") echo "Europe Central (Zurich)" ;;
        "eun1") echo "Europe North (Stockholm)" ;;
        "eus1") echo "Europe South (Milan)" ;;
        "eus2") echo "Europe South (Spain)" ;;

        # Asia Pacific
        "aps1") echo "Asia Pacific (Mumbai)" ;;
        "aps2") echo "Asia Pacific (Hyderabad)" ;;
        "apse1") echo "Asia Pacific (Singapore)" ;;
        "apse2") echo "Asia Pacific (Sydney)" ;;
        "apse3") echo "Asia Pacific (Jakarta)" ;;
        "apse4") echo "Asia Pacific (Melbourne)" ;;
        "apne1") echo "Asia Pacific (Tokyo)" ;;
        "apne2") echo "Asia Pacific (Seoul)" ;;
        "apne3") echo "Asia Pacific (Osaka)" ;;

        # Middle East
        "mec1") echo "Middle East (UAE)" ;;
        "mees1") echo "Middle East (Bahrain)" ;;

        # Africa
        "afc1") echo "Africa (Cape Town)" ;;

        # Default
        *) echo "Unknown Region" ;;
    esac
}
